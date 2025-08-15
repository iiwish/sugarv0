-- 货币资金表
CREATE TABLE `db_cash_and_equivalents` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `bank_name` VARCHAR(255) NOT NULL COMMENT '银行名称',
  `account_type` VARCHAR(100) COMMENT '账户类型',
  `currency` VARCHAR(50) COMMENT '币种',
  `beginning_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年初金额',
  `ending_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年末金额'
) COMMENT='货币资金表';

-- 应收账款表
CREATE TABLE `db_accounts_receivable` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `customer_name` VARCHAR(255) NOT NULL COMMENT '客户名称',
  `aging` VARCHAR(50) COMMENT '账龄',
  `business_type` VARCHAR(100) COMMENT '业务类型',
  `beginning_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年初金额',
  `ending_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年末金额'
) COMMENT='应收账款表';

-- 存货表
CREATE TABLE `db_inventory` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `inventory_type` VARCHAR(100) NOT NULL COMMENT '存货类型',
  `warehouse` VARCHAR(255) COMMENT '仓库',
  `inventory_status` VARCHAR(100) COMMENT '存货状态',
  `beginning_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年初金额',
  `ending_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年末金额'
) COMMENT='存货表';

-- 固定资产表
CREATE TABLE `db_fixed_assets` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `asset_type` VARCHAR(100) NOT NULL COMMENT '资产类型',
  `department` VARCHAR(100) COMMENT '使用部门',
  `depreciation_period` INT COMMENT '折旧年限（年）',
  `beginning_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年初金额',
  `ending_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年末金额'
) COMMENT='固定资产表';

-- 无形资产表
CREATE TABLE `db_intangible_assets` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `asset_type` VARCHAR(100) NOT NULL COMMENT '资产类型',
  `acquisition_method` VARCHAR(100) COMMENT '取得方式',
  `amortization_period` INT COMMENT '摊销年限（年）',
  `beginning_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年初金额',
  `ending_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年末金额'
) COMMENT='无形资产表';

-- 应付账款表
CREATE TABLE `db_accounts_payable` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `supplier_name` VARCHAR(255) NOT NULL COMMENT '供应商名称',
  `aging` VARCHAR(50) COMMENT '账龄',
  `purchase_type` VARCHAR(100) COMMENT '采购类型',
  `beginning_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年初金额',
  `ending_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年末金额'
) COMMENT='应付账款表';

-- 短期借款表
CREATE TABLE `db_short_term_loans` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `lending_bank` VARCHAR(255) NOT NULL COMMENT '贷款银行',
  `loan_purpose` TEXT COMMENT '借款用途',
  `interest_rate_range` VARCHAR(50) COMMENT '利率区间',
  `beginning_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年初金额',
  `ending_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年末金额'
) COMMENT='短期借款表';

-- 长期借款表
CREATE TABLE `db_long_term_loans` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `lending_bank` VARCHAR(255) NOT NULL COMMENT '贷款银行',
  `loan_purpose` TEXT COMMENT '借款用途',
  `interest_rate_range` VARCHAR(50) COMMENT '利率区间',
  `beginning_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年初金额',
  `ending_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年末金额'
) COMMENT='长期借款表';

-- 实收资本表
CREATE TABLE `db_paid_in_capital` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `shareholder_type` VARCHAR(100) COMMENT '股东类型',
  `contribution_method` VARCHAR(100) COMMENT '出资方式',
  `shareholder_name` VARCHAR(255) NOT NULL COMMENT '股东名称',
  `beginning_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年初金额',
  `ending_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年末金额'
) COMMENT='实收资本表';

-- 未分配利润表
CREATE TABLE `db_retained_earnings` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `year` YEAR NOT NULL COMMENT '年度',
  `profit_source` VARCHAR(255) COMMENT '利润来源',
  `distribution_status` VARCHAR(255) COMMENT '分配情况',
  `beginning_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年初金额',
  `ending_balance` DECIMAL(18, 2) DEFAULT 0.00 COMMENT '年末金额'
) COMMENT='未分配利润表';