CREATE TABLE `blacklist_config`
(
    `id`             INT UNSIGNED PRIMARY KEY,

    `customer_count` INT UNSIGNED NOT NULL,
    `merchant_count` INT UNSIGNED NOT NULL,
    `terminal_count` INT UNSIGNED NOT NULL,

    `created_at`     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE `blacklist_customer`
(
    `hash`       INT UNSIGNED NOT NULL,
    `phone`      VARCHAR(20) NOT NULL,
    `status`     SMALLINT UNSIGNED NOT NULL,
    `start_time` DATETIME NULL,
    `end_time`   DATETIME NULL,

    `created_at` TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (`hash`, `phone`)
);

CREATE TABLE `blacklist_merchant`
(
    `hash`          INT UNSIGNED NOT NULL,
    `merchant_code` VARCHAR(30) NOT NULL,
    `status`        SMALLINT UNSIGNED NOT NULL,
    `start_time`    DATETIME NULL,
    `end_time`      DATETIME NULL,

    `created_at`    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (`hash`, `merchant_code`)
);

CREATE TABLE `blacklist_terminal`
(
    `hash`          INT UNSIGNED NOT NULL,
    `merchant_code` VARCHAR(30) NOT NULL,
    `terminal_code` VARCHAR(30) NOT NULL,
    `status`        SMALLINT UNSIGNED NOT NULL,
    `start_time`    DATETIME NULL,
    `end_time`      DATETIME NULL,

    `created_at`    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (`hash`, `merchant_code`, `terminal_code`)
);

CREATE TABLE `campaign`
(
    `id`                        INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `name`                      VARCHAR(100) NOT NULL,
    `status`                    SMALLINT UNSIGNED NOT NULL,
    `type`                      SMALLINT UNSIGNED NOT NULL,

    `voucher_hash`              INT UNSIGNED NOT NULL,
    `voucher_code`              VARCHAR(30)  NOT NULL,
    `start_time`                DATETIME     NOT NULL,
    `end_time`                  DATETIME     NOT NULL,

    `budget_max`                DECIMAL(19, 2) NULL,
    `campaign_usage_max`        INT UNSIGNED NULL,
    `customer_usage_max`        INT UNSIGNED NOT NULL,

    `period_usage_type`         SMALLINT UNSIGNED NOT NULL,
    `period_customer_usage_max` INT UNSIGNED NULL,
    `period_term_type`          SMALLINT UNSIGNED NOT NULL,

    `all_merchants`             TINYINT UNSIGNED NOT NULL,

    `created_at`                TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`                TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX                       `idx_voucher_end_time` (`voucher_hash`, `voucher_code`, `end_time`)
);

CREATE TABLE `campaign_bank`
(
    `campaign_id` INT UNSIGNED NOT NULL,
    `hash`        INT UNSIGNED NOT NULL,
    `bank_code`   VARCHAR(20) NOT NULL,

    `status`      SMALLINT UNSIGNED NOT NULL,

    `created_at`  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (`campaign_id`, `hash`, `bank_code`)
);

CREATE TABLE `campaign_benefit`
(
    `id`                  INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `campaign_id`         INT UNSIGNED NOT NULL,

    `start_time`          DATETIME       NOT NULL,
    `end_time`            DATETIME       NOT NULL,

    `txn_min_amount`      DECIMAL(19, 2) NOT NULL,
    `discount_percent`    DECIMAL(19, 2) NOT NULL,
    `max_discount_amount` DECIMAL(19, 2) NOT NULL,

    `created_at`          TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`          TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX                 `idx_campaign_id_end_time` (`campaign_id`, `end_time`)
);

CREATE TABLE `campaign_customer`
(
    `campaign_id` INT UNSIGNED NOT NULL,
    `hash`        INT UNSIGNED NOT NULL,
    `phone`       VARCHAR(20) NOT NULL,

    `status`      SMALLINT UNSIGNED NOT NULL,
    `start_time`  DATETIME NULL,
    `end_time`    DATETIME NULL,

    `created_at`  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (`campaign_id`, `hash`, `phone`)
);

CREATE TABLE `campaign_merchant`
(
    `campaign_id`   INT UNSIGNED NOT NULL,
    `hash`          INT UNSIGNED NOT NULL,
    `merchant_code` VARCHAR(30) NOT NULL,

    `status`        SMALLINT UNSIGNED NOT NULL,
    `start_time`    DATETIME NULL,
    `end_time`      DATETIME NULL,
    `all_terminals` TINYINT UNSIGNED NOT NULl,

    `created_at`    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (`campaign_id`, `hash`, `merchant_code`)
);

CREATE TABLE `campaign_terminal`
(
    `campaign_id`   INT UNSIGNED NOT NULL,
    `hash`          INT UNSIGNED NOT NULL,
    `merchant_code` VARCHAR(30) NOT NULL,
    `terminal_code` VARCHAR(30) NOT NULL,

    `status`        SMALLINT UNSIGNED NOT NULL,
    `start_time`    DATETIME NULL,
    `end_time`      DATETIME NULL,

    `created_at`    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (`campaign_id`, `hash`, `merchant_code`)
);

CREATE TABLE `campaign_usage`
(
    `campaign_id`   INT UNSIGNED PRIMARY KEY,
    `budget_used`   DECIMAL(19, 2) NOT NULL,
    `campaign_used` INT UNSIGNED NOT NULL,

    `created_at`    TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE `campaign_period_usage`
(
    `campaign_id` INT UNSIGNED NOT NULL,
    `hash`        INT UNSIGNED NOT NULL,
    `phone`       VARCHAR(20)  NOT NULL,
    `term_code`   VARCHAR(100) NOT NULL,

    `usage_num`   INT UNSIGNED NOT NULL,
    `expired_on`  DATETIME     NOT NULL,

    `created_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (`campaign_id`, `hash`, `phone`, `term_code`)
);

CREATE TABLE `event`
(
    `id`             BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `seq`            BIGINT UNSIGNED NULL,
    `data`           BLOB      NOT NULL,
    `aggregate_type` SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    `aggregate_id`   INT UNSIGNED NOT NULL DEFAULT 0,
    `created_at`     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY `uk_seq` (`seq`),
    INDEX            `idx_aggregate` (`aggregate_type`, `aggregate_id`)
);