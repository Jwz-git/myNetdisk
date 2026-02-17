# 个人网盘-Go语言练手项目

## 项目配置
### 1. 需新建 uploads 文件夹
### 2. 数据库
```
mysql.server start
mysql -u root -p  
```
```sql
CREATE DATABASE personal_disk CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE personal_disk;
CREATE TABLE `file_info` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `file_name` VARCHAR(255) NOT NULL,
    `file_path` VARCHAR(500) NOT NULL,
    `file_size` BIGINT,
    `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 项目启动
### 1. 启动数据库和服务端
```
mysql.server start
go run .
```
### 2. 访问
```
访客页面
http://localhost:8080/

管理员页面
http://localhost:8080/admin
```