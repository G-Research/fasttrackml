CREATE TABLE `namespaces` (
    `id` integer,
    `code` text NOT NULL,
    `description` text,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime,
    `default_experiment_id` integer NOT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `uni_namespaces_code` UNIQUE (`code`)
);

CREATE INDEX `idx_namespaces_deleted_at` ON `namespaces`(`deleted_at`);

CREATE INDEX `idx_namespaces_code` ON `namespaces`(`code`);

CREATE TABLE `experiments` (
    `experiment_id` integer NOT NULL,
    `name` varchar(256) NOT NULL,
    `artifact_location` varchar(256),
    `lifecycle_stage` varchar(32),
    `creation_time` bigint,
    `last_update_time` bigint,
    `namespace_id` integer NOT NULL,
    PRIMARY KEY (`experiment_id`),
    CONSTRAINT `fk_namespaces_experiments` FOREIGN KEY (`namespace_id`) REFERENCES `namespaces`(`id`) ON DELETE CASCADE,
    CONSTRAINT `chk_experiments_lifecycle_stage` CHECK (lifecycle_stage IN ('active', 'deleted'))
);

CREATE UNIQUE INDEX `idx_experiments_name` ON `experiments`(`name`, `namespace_id`);

CREATE TABLE `experiment_tags` (
    `key` varchar(250) NOT NULL,
    `value` varchar(5000),
    `experiment_id` integer NOT NULL,
    PRIMARY KEY (`key`, `experiment_id`),
    CONSTRAINT `fk_experiments_tags` FOREIGN KEY (`experiment_id`) REFERENCES `experiments`(`experiment_id`) ON DELETE CASCADE
);

CREATE TABLE `runs` (
    `run_uuid` varchar(32) NOT NULL,
    `name` varchar(250),
    `source_type` varchar(20),
    `source_name` varchar(500),
    `entry_point_name` varchar(50),
    `user_id` varchar(256),
    `status` varchar(9),
    `start_time` bigint,
    `end_time` bigint,
    `source_version` varchar(50),
    `lifecycle_stage` varchar(20),
    `artifact_uri` varchar(200),
    `experiment_id` integer,
    `deleted_time` bigint,
    `row_num` bigint,
    PRIMARY KEY (`run_uuid`),
    CONSTRAINT `fk_experiments_runs` FOREIGN KEY (`experiment_id`) REFERENCES `experiments`(`experiment_id`) ON DELETE CASCADE,
    CONSTRAINT `chk_runs_lifecycle_stage` CHECK (lifecycle_stage IN ('active', 'deleted')),
    CONSTRAINT `chk_runs_source_type` CHECK (
        source_type IN ('NOTEBOOK', 'JOB', 'LOCAL', 'UNKNOWN', 'PROJECT')
    ),
    CONSTRAINT `chk_runs_status` CHECK (
        status IN (
            'SCHEDULED',
            'FAILED',
            'FINISHED',
            'RUNNING',
            'KILLED'
        )
    )
);

CREATE INDEX `idx_runs_row_num` ON `runs`(`row_num`);

CREATE TABLE `params` (
    `key` varchar(250) NOT NULL,
    `value` varchar(500) NOT NULL,
    `run_uuid` varchar(32) NOT NULL,
    PRIMARY KEY (`key`, `run_uuid`),
    CONSTRAINT `fk_runs_params` FOREIGN KEY (`run_uuid`) REFERENCES `runs`(`run_uuid`) ON DELETE CASCADE
);

CREATE INDEX `idx_params_run_id` ON `params`(`run_uuid`);

CREATE TABLE `tags` (
    `key` varchar(250) NOT NULL,
    `value` varchar(5000),
    `run_uuid` varchar(32) NOT NULL,
    PRIMARY KEY (`key`, `run_uuid`),
    CONSTRAINT `fk_runs_tags` FOREIGN KEY (`run_uuid`) REFERENCES `runs`(`run_uuid`) ON DELETE CASCADE
);

CREATE INDEX `idx_tags_run_id` ON `tags`(`run_uuid`);

CREATE TABLE `contexts` (
    `id` integer,
    `json` JSONB NOT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `uni_contexts_json` UNIQUE (`json`)
);

CREATE INDEX `idx_contexts_json` ON `contexts`(`json`);

CREATE TABLE `metrics` (
    `key` varchar(250) NOT NULL,
    `value` double precision NOT NULL,
    `timestamp` integer NOT NULL,
    `run_uuid` varchar(32) NOT NULL,
    `step` integer NOT NULL DEFAULT 0,
    `is_nan` numeric NOT NULL DEFAULT false,
    `iter` integer,
    `context_id` integer NOT NULL,
    PRIMARY KEY (
        `key`,
        `value`,
        `timestamp`,
        `run_uuid`,
        `step`,
        `is_nan`,
        `context_id`
    ),
    CONSTRAINT `fk_runs_metrics` FOREIGN KEY (`run_uuid`) REFERENCES `runs`(`run_uuid`) ON DELETE CASCADE,
    CONSTRAINT `fk_metrics_context` FOREIGN KEY (`context_id`) REFERENCES `contexts`(`id`)
);

CREATE INDEX `idx_metrics_iter` ON `metrics`(`iter`);

CREATE INDEX `idx_metrics_run_id` ON `metrics`(`run_uuid`);

CREATE TABLE `latest_metrics` (
    `key` varchar(250) NOT NULL,
    `value` double precision NOT NULL,
    `timestamp` integer,
    `step` integer NOT NULL,
    `is_nan` numeric NOT NULL,
    `run_uuid` varchar(32) NOT NULL,
    `last_iter` integer,
    `context_id` integer NOT NULL,
    PRIMARY KEY (`key`, `run_uuid`, `context_id`),
    CONSTRAINT `fk_latest_metrics_context` FOREIGN KEY (`context_id`) REFERENCES `contexts`(`id`),
    CONSTRAINT `fk_runs_latest_metrics` FOREIGN KEY (`run_uuid`) REFERENCES `runs`(`run_uuid`) ON DELETE CASCADE
);

CREATE INDEX `idx_latest_metrics_run_id` ON `latest_metrics`(`run_uuid`);

CREATE TABLE `alembic_version` (
    `version_num` varchar(32) NOT NULL,
    PRIMARY KEY (`version_num`)
);

CREATE TABLE `apps` (
    `id` uuid,
    `created_at` datetime,
    `updated_at` datetime,
    `is_archived` numeric,
    `type` text NOT NULL,
    `state` text,
    `namespace_id` integer NOT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_namespaces_apps` FOREIGN KEY (`namespace_id`) REFERENCES `namespaces`(`id`) ON DELETE CASCADE
);

CREATE TABLE `dashboards` (
    `id` uuid,
    `created_at` datetime,
    `updated_at` datetime,
    `is_archived` numeric,
    `name` text,
    `description` text,
    `app_id` uuid,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_dashboards_app` FOREIGN KEY (`app_id`) REFERENCES `apps`(`id`)
);

CREATE TABLE `schema_version` (`version` text NOT NULL, PRIMARY KEY (`version`));

CREATE TABLE `registered_models` (
    `name` varchar(256) NOT NULL,
    `creation_time` bigint NOT NULL,
    `last_updated_time` bigint NOT NULL,
    `description` varchar(5000),
    PRIMARY KEY (`name`)
);

CREATE TABLE `model_versions` (
    `name` varchar(256) NOT NULL,
    `version` integer NOT NULL,
    `creation_time` bigint NOT NULL,
    `last_updated_time` bigint NOT NULL,
    `description` varchar(5000),
    `user_id` varchar(256),
    `current_stage` varchar(20),
    `source` varchar(500),
    `run_id` varchar(32),
    `status` varchar(20),
    `status_message` varchar(500),
    `run_link` varchar(500),
    PRIMARY KEY (`name`, `version`),
    CONSTRAINT `fk_model_versions_registered_model` FOREIGN KEY (`name`) REFERENCES `registered_models`(`name`) ON UPDATE CASCADE
);

CREATE TABLE `model_version_tags` (
    `key` varchar(250) NOT NULL,
    `value` varachar(5000),
    `name` varchar(256) NOT NULL,
    `version` integer NOT NULL,
    PRIMARY KEY (`key`, `name`, `version`),
    CONSTRAINT `fk_model_version_tags_model_version` FOREIGN KEY (`name`, `version`) REFERENCES `model_versions`(`name`, `version`) ON UPDATE CASCADE
);

CREATE TABLE `registered_model_tags` (
    `key` varchar(250) NOT NULL,
    `value` varachar(5000),
    `name` varchar(256) NOT NULL,
    `registered_model_name` varchar(256) NOT NULL,
    PRIMARY KEY (`key`, `name`),
    CONSTRAINT `fk_registered_model_tags_registered_model` FOREIGN KEY (`registered_model_name`) REFERENCES `registered_models`(`name`) ON UPDATE CASCADE
);

CREATE TABLE `registered_model_aliases` (
    `alias` varchar(256) NOT NULL,
    `version` integer NOT NULL,
    `name` varchar(256) NOT NULL,
    PRIMARY KEY (`alias`, `name`),
    CONSTRAINT `fk_registered_model_aliases_registered_model` FOREIGN KEY (`name`) REFERENCES `registered_models`(`name`) ON DELETE CASCADE ON UPDATE CASCADE
);