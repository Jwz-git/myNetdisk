const MAX_UINT16 = 0xffff;
const MAX_UINT32 = 0xffffffff;
const UTF8_FLAG = 0x0800;
const STORE_METHOD = 0;
const DEFLATE_METHOD = 8;

const textEncoder = new TextEncoder();
const crcTable = new Uint32Array(256);

for (let value = 0; value < crcTable.length; value += 1) {
    let remainder = value;
    for (let bit = 0; bit < 8; bit += 1) {
        remainder = (remainder & 1) !== 0
            ? 0xedb88320 ^ (remainder >>> 1)
            : remainder >>> 1;
    }
    crcTable[value] = remainder >>> 0;
}

const crc32 = (bytes) => {
    let checksum = 0xffffffff;
    for (const byte of bytes) {
        checksum = crcTable[(checksum ^ byte) & 0xff] ^ (checksum >>> 8);
    }
    return (checksum ^ 0xffffffff) >>> 0;
};

const normalizeZipPath = (rawPath) => {
    const normalized = String(rawPath || '').replaceAll('\\', '/');
    if (!normalized || normalized.startsWith('/') || /^[A-Za-z]:/.test(normalized)) {
        throw new Error(`ZIP 内文件路径无效：${rawPath}`);
    }

    const parts = normalized.split('/');
    if (parts.some((part) => !part || part === '.' || part === '..' || part.includes('\0'))) {
        throw new Error(`ZIP 内文件路径无效：${rawPath}`);
    }
    return parts.join('/');
};

const getDOSDateTime = (lastModified) => {
    const parsed = new Date(Number(lastModified) || Date.now());
    const date = Number.isNaN(parsed.getTime()) ? new Date() : parsed;
    const year = Math.min(2107, Math.max(1980, date.getFullYear()));

    return {
        time: (date.getHours() << 11) | (date.getMinutes() << 5) | Math.floor(date.getSeconds() / 2),
        date: ((year - 1980) << 9) | ((date.getMonth() + 1) << 5) | date.getDate()
    };
};

const deflateRaw = async (bytes) => {
    if (typeof CompressionStream !== 'function' || bytes.length === 0) return null;

    try {
        const stream = new Blob([bytes])
            .stream()
            .pipeThrough(new CompressionStream('deflate-raw'));
        return new Uint8Array(await new Response(stream).arrayBuffer());
    } catch {
        return null;
    }
};

const prepareEntryData = async (bytes) => {
    const compressed = await deflateRaw(bytes);
    if (compressed && compressed.length < bytes.length) {
        return { method: DEFLATE_METHOD, bytes: compressed };
    }
    return { method: STORE_METHOD, bytes };
};

const createLocalHeader = (entry) => {
    const header = new Uint8Array(30 + entry.nameBytes.length);
    const view = new DataView(header.buffer);
    view.setUint32(0, 0x04034b50, true);
    view.setUint16(4, 20, true);
    view.setUint16(6, UTF8_FLAG, true);
    view.setUint16(8, entry.method, true);
    view.setUint16(10, entry.time, true);
    view.setUint16(12, entry.date, true);
    view.setUint32(14, entry.checksum, true);
    view.setUint32(18, entry.compressedSize, true);
    view.setUint32(22, entry.originalSize, true);
    view.setUint16(26, entry.nameBytes.length, true);
    view.setUint16(28, 0, true);
    header.set(entry.nameBytes, 30);
    return header;
};

const createCentralHeader = (entry) => {
    const header = new Uint8Array(46 + entry.nameBytes.length);
    const view = new DataView(header.buffer);
    view.setUint32(0, 0x02014b50, true);
    view.setUint16(4, 20, true);
    view.setUint16(6, 20, true);
    view.setUint16(8, UTF8_FLAG, true);
    view.setUint16(10, entry.method, true);
    view.setUint16(12, entry.time, true);
    view.setUint16(14, entry.date, true);
    view.setUint32(16, entry.checksum, true);
    view.setUint32(20, entry.compressedSize, true);
    view.setUint32(24, entry.originalSize, true);
    view.setUint16(28, entry.nameBytes.length, true);
    view.setUint16(30, 0, true);
    view.setUint16(32, 0, true);
    view.setUint16(34, 0, true);
    view.setUint16(36, 0, true);
    view.setUint32(38, 0, true);
    view.setUint32(42, entry.localOffset, true);
    header.set(entry.nameBytes, 46);
    return header;
};

const createEndRecord = (entryCount, centralSize, centralOffset) => {
    const record = new Uint8Array(22);
    const view = new DataView(record.buffer);
    view.setUint32(0, 0x06054b50, true);
    view.setUint16(4, 0, true);
    view.setUint16(6, 0, true);
    view.setUint16(8, entryCount, true);
    view.setUint16(10, entryCount, true);
    view.setUint32(12, centralSize, true);
    view.setUint32(16, centralOffset, true);
    view.setUint16(20, 0, true);
    return record;
};

export async function createZipBlob(files, onProgress = () => {}) {
    const sourceFiles = Array.from(files);
    if (sourceFiles.length === 0) throw new Error('文件夹中没有可上传的文件');
    if (sourceFiles.length > MAX_UINT16) throw new Error('文件夹内文件数量超过 ZIP 格式限制');

    const bodyParts = [];
    const entries = [];
    let bodyOffset = 0;

    for (let index = 0; index < sourceFiles.length; index += 1) {
        const file = sourceFiles[index];
        const name = normalizeZipPath(file.webkitRelativePath || file.name);
        const nameBytes = textEncoder.encode(name);
        if (nameBytes.length > MAX_UINT16) throw new Error(`文件路径过长：${name}`);

        const originalBytes = new Uint8Array(await file.arrayBuffer());
        if (originalBytes.length > MAX_UINT32) throw new Error(`单个文件超过 4GB：${name}`);

        const prepared = await prepareEntryData(originalBytes);
        const dos = getDOSDateTime(file.lastModified);
        const entry = {
            nameBytes,
            method: prepared.method,
            time: dos.time,
            date: dos.date,
            checksum: crc32(originalBytes),
            compressedSize: prepared.bytes.length,
            originalSize: originalBytes.length,
            localOffset: bodyOffset
        };
        const localHeader = createLocalHeader(entry);

        bodyParts.push(localHeader, prepared.bytes);
        bodyOffset += localHeader.length + prepared.bytes.length;
        if (bodyOffset > MAX_UINT32) throw new Error('ZIP 压缩包超过 4GB，当前浏览器上传不支持');

        entries.push(entry);
        onProgress({ completed: index + 1, total: sourceFiles.length, name });
    }

    const centralParts = entries.map(createCentralHeader);
    const centralSize = centralParts.reduce((sum, part) => sum + part.length, 0);
    if (bodyOffset + centralSize > MAX_UINT32) {
        throw new Error('ZIP 压缩包超过 4GB，当前浏览器上传不支持');
    }

    return new Blob(
        [...bodyParts, ...centralParts, createEndRecord(entries.length, centralSize, bodyOffset)],
        { type: 'application/zip' }
    );
}
