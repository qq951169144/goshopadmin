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
    PRIMARY KEY (role_id, permission_id),
    INDEX idx_role_permissions_permission_id (permission_id)
);

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(191) NOT NULL UNIQUE,
    password longtext NOT NULL,
    role_id INT NOT NULL,
    status enum('active','inactive') DEFAULT 'active',
    created_at datetime(3) DEFAULT NULL,
    updated_at datetime(3) DEFAULT NULL,
    email VARCHAR(191) DEFAULT NULL,
    INDEX idx_users_role_id (role_id),
    INDEX idx_users_status (status),
    UNIQUE INDEX email (email)
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
    approved_by INT DEFAULT NULL,
    INDEX idx_merchants_status (status),
    INDEX idx_merchants_audit_status (audit_status)
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
    UNIQUE KEY unique_merchant_user (merchant_id, user_id),
    INDEX idx_merchant_users_merchant_id (merchant_id),
    INDEX idx_merchant_users_user_id (user_id)
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
    audited_at DATETIME DEFAULT NULL,
    INDEX idx_merchant_audit_merchant_id (merchant_id),
    INDEX idx_merchant_audit_status (status)
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
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_merchant_bank_merchant_id (merchant_id),
    INDEX idx_merchant_bank_is_default (is_default)
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
    UNIQUE KEY unique_withdraw_order_no (order_no),
    INDEX idx_merchant_withdraw_merchant_id (merchant_id),
    INDEX idx_merchant_withdraw_status (status)
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
    UNIQUE KEY unique_statement_no (statement_no),
    INDEX idx_merchant_statement_merchant_id (merchant_id),
    INDEX idx_merchant_statement_status (status)
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
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_product_categories_merchant_id (merchant_id),
    INDEX idx_product_categories_parent_id (parent_id),
    INDEX idx_product_categories_status (status)
);

-- 创建商品表
CREATE TABLE IF NOT EXISTS products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description longtext,
    detail TEXT,
    price DECIMAL(10,2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    category_id INT NOT NULL,
    merchant_id INT NOT NULL,
    status enum('active','inactive') DEFAULT 'active',
    created_at datetime(3) DEFAULT NULL,
    updated_at datetime(3) DEFAULT NULL,
    INDEX idx_products_merchant_id (merchant_id),
    INDEX idx_products_category_id (category_id),
    INDEX idx_products_status (status)
);

-- 创建商品SKU表
CREATE TABLE IF NOT EXISTS product_skus (
    id INT AUTO_INCREMENT PRIMARY KEY,
    product_id INT NOT NULL,
    merchant_id INT NOT NULL,
    sku_code VARCHAR(50) NOT NULL,
    attributes JSON,
    price DECIMAL(10,2) NOT NULL,
    original_price DECIMAL(10,2) DEFAULT 0 COMMENT '原价',
    stock INT NOT NULL,
    status enum('active','inactive') DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX unique_sku_code (sku_code),
    INDEX idx_product_skus_product_id (product_id),
    INDEX idx_product_skus_merchant_id (merchant_id),
    INDEX idx_product_skus_status (status)
);

-- 创建商品规格表
CREATE TABLE IF NOT EXISTS product_specifications (
    id INT AUTO_INCREMENT PRIMARY KEY,
    product_id INT NOT NULL COMMENT '所属商品ID',
    name VARCHAR(50) NOT NULL COMMENT '规格名称：颜色、尺寸、版本',
    sort INT DEFAULT 0 COMMENT '排序，控制前端展示顺序',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_product_specifications_product_id (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 创建规格值表
CREATE TABLE IF NOT EXISTS product_specification_values (
    id INT AUTO_INCREMENT PRIMARY KEY,
    spec_id INT NOT NULL COMMENT '关联规格ID',
    value VARCHAR(50) NOT NULL COMMENT '规格值：红色、M码',
    sort INT DEFAULT 0 COMMENT '排序',
    status enum('active','inactive') DEFAULT 'active' COMMENT '状态：active-启用，inactive-禁用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_specification_values_spec_id (spec_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 创建SKU规格关联表
CREATE TABLE IF NOT EXISTS product_sku_specs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    sku_id INT NOT NULL COMMENT 'SKU ID',
    spec_id INT NOT NULL COMMENT '规格ID',
    spec_value_id INT NOT NULL COMMENT '规格值ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_sku_specs_sku_id (sku_id),
    INDEX idx_sku_specs_spec_id (spec_id),
    INDEX idx_sku_specs_spec_value_id (spec_value_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 创建商品图片表
CREATE TABLE IF NOT EXISTS product_images (
    id INT AUTO_INCREMENT PRIMARY KEY,
    product_id INT NOT NULL,
    image_url VARCHAR(255) NOT NULL,
    is_main TINYINT(1) DEFAULT 0,
    sort INT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_product_images_product_id (product_id),
    INDEX idx_product_images_is_main (is_main)
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
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    nickname VARCHAR(50) DEFAULT NULL,
    avatar VARCHAR(255) DEFAULT NULL,
    last_login_at DATETIME DEFAULT NULL,
    last_login_ip VARCHAR(50) DEFAULT NULL,
    INDEX idx_customers_phone (phone)
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
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_addresses_customer_id (customer_id),
    INDEX idx_addresses_is_default (is_default)
);

-- 创建订单表
CREATE TABLE IF NOT EXISTS orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_no VARCHAR(32) NOT NULL UNIQUE,
    customer_id INT NOT NULL,
    merchant_id INT NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(191) NOT NULL DEFAULT 'pending',
    address_id INT NOT NULL,
    created_at datetime(3) DEFAULT NULL,
    updated_at datetime(3) DEFAULT NULL,
    payment_method longtext,
    transaction_id longtext,
    paid_at DATETIME DEFAULT NULL,
    shipped_at DATETIME DEFAULT NULL,
    delivered_at DATETIME DEFAULT NULL,
    cancelled_at DATETIME DEFAULT NULL,
    INDEX idx_orders_customer_id (customer_id),
    INDEX idx_orders_merchant_id (merchant_id),
    INDEX idx_orders_address_id (address_id),
    INDEX idx_orders_status (status)
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
    created_at datetime(3) DEFAULT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_order_items_order_id (order_id),
    INDEX idx_order_items_product_id (product_id)
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
    UNIQUE KEY unique_payment_no (payment_no),
    INDEX idx_payments_order_id (order_id),
    INDEX idx_payments_status (status)
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
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_activities_merchant_id (merchant_id),
    INDEX idx_activities_created_by (created_by)
);

-- 创建活动规则表
CREATE TABLE IF NOT EXISTS activity_rules (
    id INT AUTO_INCREMENT PRIMARY KEY,
    activity_id INT NOT NULL,
    rule_type VARCHAR(20) NOT NULL,
    rule_value JSON NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_activity_rules_activity_id (activity_id)
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
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_activity_products_activity_id (activity_id),
    INDEX idx_activity_products_product_id (product_id),
    INDEX idx_activity_products_merchant_id (merchant_id)
);

-- 创建活动效果统计表
CREATE TABLE IF NOT EXISTS activity_stats (
    id INT AUTO_INCREMENT PRIMARY KEY,
    activity_id INT NOT NULL,
    view_count INT DEFAULT 0,
    participant_count INT DEFAULT 0,
    order_count INT DEFAULT 0,
    total_amount DECIMAL(10,2) DEFAULT 0.00,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_activity_stats_activity_id (activity_id)
);

-- 创建物流表
CREATE TABLE IF NOT EXISTS shipping (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    tracking_no VARCHAR(50),
    status VARCHAR(20) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_shipping_order_id (order_id)
);

-- 创建购物车表
CREATE TABLE IF NOT EXISTS carts (
    id INT NOT NULL AUTO_INCREMENT,
    user_id INT NOT NULL DEFAULT 0,
    session_id VARCHAR(255) DEFAULT NULL,
    created_at datetime(3) DEFAULT NULL,
    updated_at datetime(3) DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    INDEX `idx_carts_user_id` (`user_id`),
    INDEX `idx_carts_session_id` (`session_id`),
    INDEX `idx_carts_user_session` (`user_id`, `session_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 创建购物车项表
CREATE TABLE IF NOT EXISTS cart_items (
    id INT NOT NULL AUTO_INCREMENT,
    cart_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL DEFAULT '1',
    price decimal(10,2) NOT NULL,
    sku VARCHAR(255) DEFAULT NULL,
    created_at datetime(3) DEFAULT NULL,
    updated_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_cart_items_cart_id` (`cart_id`),
    INDEX `idx_cart_items_product_id` (`product_id`),
    INDEX `idx_cart_items_cart_product` (`cart_id`, `product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 创建支付方式表
CREATE TABLE IF NOT EXISTS payment_methods (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    code VARCHAR(20) NOT NULL,
    status enum('active','inactive') DEFAULT 'active',
    created_at datetime(0) DEFAULT CURRENT_TIMESTAMP(0),
    updated_at datetime(0) DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
    PRIMARY KEY (`id`),
    UNIQUE INDEX `unique_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建物流方式表
CREATE TABLE IF NOT EXISTS shipping_methods (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    code VARCHAR(20) NOT NULL,
    price decimal(10,2) NOT NULL,
    status enum('active','inactive') DEFAULT 'active',
    created_at datetime(0) DEFAULT CURRENT_TIMESTAMP(0),
    updated_at datetime(0) DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
    PRIMARY KEY (`id`),
    UNIQUE INDEX `unique_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
INSERT INTO users (username, password, role_id, status, created_at, updated_at) VALUES
('admin', '$2a$10$ppQSNCrBWqi5Xde1BSc1weysPKc8bfo2bZsvEtwoHrNxzWAD7kWge', 1, 'active', '2026-03-12 11:00:10.000', '2026-03-12 11:00:10.000');

-- 插入默认平台管理员账号（密码：123456）
INSERT INTO users (username, password, role_id, status, created_at, updated_at) VALUES
('platform', '$2a$10$ppQSNCrBWqi5Xde1BSc1weysPKc8bfo2bZsvEtwoHrNxzWAD7kWge', 2, 'active', '2026-03-12 11:00:10.000', '2026-03-12 11:00:10.000');

-- 插入默认商户账号（密码：123456）
INSERT INTO users (username, password, role_id, status, created_at, updated_at) VALUES
('merchant', '$2a$10$ppQSNCrBWqi5Xde1BSc1weysPKc8bfo2bZsvEtwoHrNxzWAD7kWge', 3, 'active', '2026-03-12 11:00:10.000', '2026-03-12 11:00:10.000');

-- 插入测试商户用户
INSERT INTO users (username, password, role_id, status, created_at, updated_at) VALUES
('pxj', '$2a$10$NSDs.5eM0j1d2BfE.kw0UuDxak8.eVF.AdAl.yXnhjxypEzJ5XvDG', 3, 'active', '2026-03-12 11:00:58.797', '2026-03-12 11:00:58.797');

-- 插入测试商户
INSERT INTO merchants (id, name, contact_name, contact_phone, email, address, business_license, tax_number, audit_status, status, created_at, updated_at, approved_at, approved_by) VALUES
(1, '小商店', '法人', '18819279587', '951169144@qq.com', '广州', 'jaskdhasjk', '123456789', 'approved', 'active', '2026-03-12 11:02:32', '2026-03-12 11:02:50', '2026-03-12 11:02:50', 1);

-- 插入商户用户关联
INSERT INTO merchant_users (id, merchant_id, user_id, `role`, status, created_at, updated_at) VALUES
(1, 1, 4, 'owner', 'active', '2026-03-12 11:02:39', '2026-03-12 11:02:39');

-- 插入商户审核记录
INSERT INTO merchant_audit (id, merchant_id, audit_type, old_data, new_data, status, remark, created_by, audited_by, created_at, audited_at) VALUES
(1, 1, 'registration', '{}', '{"name": "小商店", "email": "951169144@qq.com", "address": "广州", "tax_number": "123456789", "contact_name": "法人", "contact_phone": "18819279587", "business_license": "jaskdhasjk"}', 'approved', '1', 1, 1, '2026-03-12 11:02:32', '2026-03-12 11:02:50');

-- 插入商品分类
INSERT INTO product_categories (id, merchant_id, name, parent_id, level, sort, status, created_at, updated_at) VALUES
(1, 1, '日用品', 0, 1, 1, 'active', '2026-03-12 11:03:13', '2026-03-12 11:03:13'),
(2, 1, '小玩具', 0, 1, 2, 'active', '2026-03-12 11:03:19', '2026-03-12 11:03:19'),
(3, 1, '食物', 0, 1, 0, 'active', '2026-03-12 11:03:27', '2026-03-12 11:03:27'),
(4, 1, '虚拟商品', 0, 1, 0, 'active', '2026-03-12 11:03:32', '2026-03-12 11:03:32');

-- 提交事务
COMMIT;