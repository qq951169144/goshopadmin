-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS goshopadmin CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE goshopadmin;

-- 创建角色表
CREATE TABLE IF NOT EXISTS roles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(200),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建权限表
CREATE TABLE IF NOT EXISTS permissions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description VARCHAR(200),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INT NOT NULL,
    permission_id INT NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL,
    role_id INT NOT NULL,
    status enum('active','inactive') DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建商品表
CREATE TABLE IF NOT EXISTS products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    stock INT NOT NULL,
    category VARCHAR(50),
    merchant_id INT NOT NULL,
    status enum('active','inactive') DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建订单表
CREATE TABLE IF NOT EXISTS orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_no VARCHAR(32) NOT NULL UNIQUE,
    user_id INT NOT NULL,
    merchant_id INT NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建活动表
CREATE TABLE IF NOT EXISTS activities (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    status enum('active','inactive') DEFAULT 'active',
    created_by INT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建物流表
CREATE TABLE IF NOT EXISTS shipping (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    tracking_no VARCHAR(50),
    status VARCHAR(20) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 插入默认角色
INSERT INTO roles (name, description) VALUES
('超级管理员', '系统最高权限'),
('平台管理员', '平台管理权限'),
('商户账号', '商户管理权限');

-- 插入默认权限
INSERT INTO permissions (name, code, description) VALUES
('用户管理', 'user:manage', '管理用户账号'),
('角色管理', 'role:manage', '管理角色权限'),
('商品管理', 'product:manage', '管理商品信息'),
('订单管理', 'order:manage', '管理订单信息'),
('活动管理', 'activity:manage', '管理活动信息'),
('物流管理', 'shipping:manage', '管理物流信息');

-- 为超级管理员分配所有权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5), (1, 6);

-- 为平台管理员分配部分权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
(2, 1), (2, 2), (2, 3), (2, 4), (2, 5), (2, 6);

-- 为商户账号分配部分权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
(3, 3), (3, 4), (3, 5), (3, 6);

-- 插入默认超级管理员账号（密码：123456）
INSERT INTO users (username, password, role_id, status) VALUES
('admin', '$2a$10$ppQSNCrBWqi5Xde1BSc1weysPKc8bfo2bZsvEtwoHrNxzWAD7kWge', 1, 'active');

-- 插入默认平台管理员账号（密码：123456）
INSERT INTO users (username, password, role_id, status) VALUES
('platform', '$2a$10$ppQSNCrBWqi5Xde1BSc1weysPKc8bfo2bZsvEtwoHrNxzWAD7kWge', 2, 'active');

-- 插入默认商户账号（密码：123456）
INSERT INTO users (username, password, role_id, status) VALUES
('merchant', '$2a$10$ppQSNCrBWqi5Xde1BSc1weysPKc8bfo2bZsvEtwoHrNxzWAD7kWge', 3, 'active');

-- 创建索引
CREATE INDEX idx_users_role_id ON users(role_id);
CREATE INDEX idx_products_merchant_id ON products(merchant_id);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_merchant_id ON orders(merchant_id);
CREATE INDEX idx_activities_created_by ON activities(created_by);
CREATE INDEX idx_shipping_order_id ON shipping(order_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- 提交事务
COMMIT;