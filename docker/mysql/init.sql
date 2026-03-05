-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS goshopadmin CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE goshopadmin;

-- 创建角色表
CREATE TABLE IF NOT EXISTS roles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(200),
    status enum('active','inactive') DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建权限表
CREATE TABLE IF NOT EXISTS permissions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description VARCHAR(200),
    status enum('active','inactive') DEFAULT 'active',
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

-- 创建商户表
CREATE TABLE IF NOT EXISTS merchants (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    contact_name VARCHAR(50) NOT NULL,
    contact_phone VARCHAR(20) NOT NULL,
    email VARCHAR(100) NOT NULL,
    address VARCHAR(255) NOT NULL,
    business_license VARCHAR(255) NOT NULL,
    tax_number VARCHAR(50) NOT NULL,
    audit_status enum('pending','approved','rejected') DEFAULT 'pending',
    status enum('active','inactive') DEFAULT 'inactive',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    approved_at DATETIME DEFAULT NULL,
    approved_by INT DEFAULT NULL
);

-- 创建商户用户关联表
CREATE TABLE IF NOT EXISTS merchant_users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    merchant_id INT NOT NULL,
    user_id INT NOT NULL,
    `role` enum('owner','manager','staff') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'owner',
    status enum('active','inactive') DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_merchant_user (merchant_id, user_id)
);

-- 创建商户审核表
CREATE TABLE IF NOT EXISTS merchant_audit (
    id INT AUTO_INCREMENT PRIMARY KEY,
    merchant_id INT NOT NULL,
    audit_type VARCHAR(20) NOT NULL,
    old_data JSON DEFAULT NULL,
    new_data JSON DEFAULT NULL,
    status enum('pending','approved','rejected') DEFAULT 'pending',
    remark TEXT DEFAULT NULL,
    created_by INT NOT NULL,
    audited_by INT DEFAULT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    audited_at DATETIME DEFAULT NULL
);

-- 创建商户银行信息表
CREATE TABLE IF NOT EXISTS merchant_bank (
    id INT AUTO_INCREMENT PRIMARY KEY,
    merchant_id INT NOT NULL,
    bank_name VARCHAR(100) NOT NULL,
    account_name VARCHAR(100) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    branch VARCHAR(100) NOT NULL,
    is_default TINYINT(1) DEFAULT 0,
    status enum('active','inactive') DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建商户提现记录表
CREATE TABLE IF NOT EXISTS merchant_withdraw (
    id INT AUTO_INCREMENT PRIMARY KEY,
    merchant_id INT NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    bank_id INT NOT NULL,
    status enum('pending','processing','completed','failed') DEFAULT 'pending',
    order_no VARCHAR(32) NOT NULL,
    remark TEXT DEFAULT NULL,
    created_by INT NOT NULL,
    processed_by INT DEFAULT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    processed_at DATETIME DEFAULT NULL,
    UNIQUE KEY unique_withdraw_order_no (order_no)
);

-- 创建商户对账单表
CREATE TABLE IF NOT EXISTS merchant_statement (
    id INT AUTO_INCREMENT PRIMARY KEY,
    merchant_id INT NOT NULL,
    statement_no VARCHAR(32) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    order_count INT NOT NULL,
    status enum('draft','confirmed','settled') DEFAULT 'draft',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_statement_no (statement_no)
);

-- 创建商品分类表
CREATE TABLE IF NOT EXISTS product_categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    merchant_id INT NOT NULL,
    name VARCHAR(50) NOT NULL,
    parent_id INT DEFAULT 0,
    level INT DEFAULT 1,
    sort INT DEFAULT 0,
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

-- 创建商品SKU表
CREATE TABLE IF NOT EXISTS product_skus (
    id INT AUTO_INCREMENT PRIMARY KEY,
    product_id INT NOT NULL,
    merchant_id INT NOT NULL,
    sku_code VARCHAR(50) NOT NULL,
    attributes JSON NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    stock INT NOT NULL,
    status enum('active','inactive') DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建商品图片表
CREATE TABLE IF NOT EXISTS product_images (
    id INT AUTO_INCREMENT PRIMARY KEY,
    product_id INT NOT NULL,
    image_url VARCHAR(255) NOT NULL,
    is_main TINYINT(1) DEFAULT 0,
    sort INT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建C端用户表
CREATE TABLE IF NOT EXISTS customers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL UNIQUE,
    email VARCHAR(100) UNIQUE,
    status enum('active','inactive') DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建地址表
CREATE TABLE IF NOT EXISTS addresses (
    id INT AUTO_INCREMENT PRIMARY KEY,
    customer_id INT NOT NULL,
    name VARCHAR(50) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    province VARCHAR(50) NOT NULL,
    city VARCHAR(50) NOT NULL,
    district VARCHAR(50) NOT NULL,
    detail_address VARCHAR(255) NOT NULL,
    is_default TINYINT(1) DEFAULT 0,
    status enum('active','inactive') DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建订单表
CREATE TABLE IF NOT EXISTS orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_no VARCHAR(32) NOT NULL UNIQUE,
    customer_id INT NOT NULL,
    merchant_id INT NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) NOT NULL,
    address_id INT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建订单明细表
CREATE TABLE IF NOT EXISTS order_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    product_id INT NOT NULL,
    sku_id INT DEFAULT NULL,
    product_name VARCHAR(100) NOT NULL,
    sku_attributes JSON DEFAULT NULL,
    price DECIMAL(10,2) NOT NULL,
    quantity INT NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建支付记录表
CREATE TABLE IF NOT EXISTS payments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    payment_no VARCHAR(32) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    payment_method VARCHAR(20) NOT NULL,
    transaction_id VARCHAR(100) DEFAULT NULL,
    status enum('pending','success','failed') DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    paid_at DATETIME DEFAULT NULL,
    UNIQUE KEY unique_payment_no (payment_no)
);

-- 创建活动表
CREATE TABLE IF NOT EXISTS activities (
    id INT AUTO_INCREMENT PRIMARY KEY,
    merchant_id INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    status enum('active','inactive') DEFAULT 'active',
    created_by INT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建活动规则表
CREATE TABLE IF NOT EXISTS activity_rules (
    id INT AUTO_INCREMENT PRIMARY KEY,
    activity_id INT NOT NULL,
    rule_type VARCHAR(20) NOT NULL,
    rule_value JSON NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建活动商品表
CREATE TABLE IF NOT EXISTS activity_products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    activity_id INT NOT NULL,
    product_id INT NOT NULL,
    sku_id INT DEFAULT NULL,
    merchant_id INT NOT NULL,
    original_price DECIMAL(10,2) NOT NULL,
    activity_price DECIMAL(10,2) NOT NULL,
    stock INT NOT NULL,
    status enum('active','inactive') DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建活动效果统计表
CREATE TABLE IF NOT EXISTS activity_stats (
    id INT AUTO_INCREMENT PRIMARY KEY,
    activity_id INT NOT NULL,
    view_count INT DEFAULT 0,
    participant_count INT DEFAULT 0,
    order_count INT DEFAULT 0,
    total_amount DECIMAL(10,2) DEFAULT 0,
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
('物流管理', 'shipping:manage', '管理物流信息'),
('商户管理', 'merchant:manage', '管理商户信息'),
('商户审核', 'merchant:audit', '审核商户注册和信息更新'),
('商户财务', 'merchant:finance', '管理商户提现和对账'),
('商户查看', 'merchant:view', '查看商户信息'),
('商品分类管理', 'product:category', '管理商品分类'),
('商品SKU管理', 'product:sku', '管理商品SKU'),
('支付管理', 'order:payment', '管理订单支付'),
('活动统计', 'activity:stats', '查看活动统计数据');

-- 为超级管理员分配所有权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5), (1, 6),
(1, 7), (1, 8), (1, 9), (1, 10), (1, 11), (1, 12), (1, 13), (1, 14);

-- 为平台管理员分配部分权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
(2, 1), (2, 2), (2, 3), (2, 4), (2, 5), (2, 6),
(2, 7), (2, 8), (2, 9), (2, 10), (2, 11), (2, 12), (2, 13), (2, 14);

-- 为商户账号分配部分权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
(3, 3), (3, 4), (3, 5), (3, 6),
(3, 10), (3, 11), (3, 12), (3, 13);

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
CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_orders_merchant_id ON orders(merchant_id);
CREATE INDEX idx_orders_address_id ON orders(address_id);
CREATE INDEX idx_shipping_order_id ON shipping(order_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX idx_merchants_status ON merchants(status);
CREATE INDEX idx_merchants_audit_status ON merchants(audit_status);
CREATE INDEX idx_merchant_users_merchant_id ON merchant_users(merchant_id);
CREATE INDEX idx_merchant_users_user_id ON merchant_users(user_id);
CREATE INDEX idx_merchant_audit_merchant_id ON merchant_audit(merchant_id);
CREATE INDEX idx_merchant_audit_status ON merchant_audit(status);
CREATE INDEX idx_merchant_bank_merchant_id ON merchant_bank(merchant_id);
CREATE INDEX idx_merchant_bank_is_default ON merchant_bank(is_default);
CREATE INDEX idx_merchant_withdraw_merchant_id ON merchant_withdraw(merchant_id);
CREATE INDEX idx_merchant_withdraw_status ON merchant_withdraw(status);
CREATE INDEX idx_merchant_statement_merchant_id ON merchant_statement(merchant_id);
CREATE INDEX idx_merchant_statement_status ON merchant_statement(status);
CREATE INDEX idx_product_categories_merchant_id ON product_categories(merchant_id);
CREATE INDEX idx_product_categories_parent_id ON product_categories(parent_id);
CREATE INDEX idx_product_categories_status ON product_categories(status);
CREATE INDEX idx_product_skus_product_id ON product_skus(product_id);
CREATE INDEX idx_product_skus_merchant_id ON product_skus(merchant_id);
CREATE INDEX idx_product_skus_status ON product_skus(status);
CREATE INDEX idx_product_images_product_id ON product_images(product_id);
CREATE INDEX idx_product_images_is_main ON product_images(is_main);
CREATE INDEX idx_customers_phone ON customers(phone);
CREATE INDEX idx_addresses_customer_id ON addresses(customer_id);
CREATE INDEX idx_addresses_is_default ON addresses(is_default);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);
CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_activities_merchant_id ON activities(merchant_id);
CREATE INDEX idx_activities_created_by ON activities(created_by);
CREATE INDEX idx_activity_rules_activity_id ON activity_rules(activity_id);
CREATE INDEX idx_activity_products_activity_id ON activity_products(activity_id);
CREATE INDEX idx_activity_products_product_id ON activity_products(product_id);
CREATE INDEX idx_activity_products_merchant_id ON activity_products(merchant_id);
CREATE INDEX idx_activity_stats_activity_id ON activity_stats(activity_id);

-- 提交事务
COMMIT;