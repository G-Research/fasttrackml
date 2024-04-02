CREATE TABLE alembic_version (
	version_num VARCHAR(32) NOT NULL, 
	CONSTRAINT alembic_version_pkc PRIMARY KEY (version_num)
);
CREATE TABLE registered_models (
	name VARCHAR(256) NOT NULL, 
	creation_time BIGINT, 
	last_updated_time BIGINT, 
	description VARCHAR(5000), 
	CONSTRAINT registered_model_pk PRIMARY KEY (name), 
	UNIQUE (name)
);
CREATE TABLE registered_model_tags (
	"key" VARCHAR(250) NOT NULL, 
	value VARCHAR(5000), 
	name VARCHAR(256) NOT NULL, 
	CONSTRAINT registered_model_tag_pk PRIMARY KEY ("key", name), 
	FOREIGN KEY(name) REFERENCES registered_models (name) ON UPDATE cascade
);
CREATE TABLE model_version_tags (
	"key" VARCHAR(250) NOT NULL, 
	value VARCHAR(5000), 
	name VARCHAR(256) NOT NULL, 
	version INTEGER NOT NULL, 
	CONSTRAINT model_version_tag_pk PRIMARY KEY ("key", name, version), 
	FOREIGN KEY(name, version) REFERENCES model_versions (name, version) ON UPDATE cascade
);
CREATE TABLE IF NOT EXISTS "model_versions" (
	name VARCHAR(256) NOT NULL, 
	version INTEGER NOT NULL, 
	creation_time BIGINT, 
	last_updated_time BIGINT, 
	description VARCHAR(5000), 
	user_id VARCHAR(256), 
	current_stage VARCHAR(20), 
	source VARCHAR(500), 
	run_id VARCHAR(32), 
	status VARCHAR(20), 
	status_message VARCHAR(500), 
	run_link VARCHAR(500), 
	CONSTRAINT model_version_pk PRIMARY KEY (name, version), 
	FOREIGN KEY(name) REFERENCES registered_models (name) ON UPDATE CASCADE
);
CREATE TABLE registered_model_aliases (
	alias VARCHAR(256) NOT NULL, 
	version INTEGER NOT NULL, 
	name VARCHAR(256) NOT NULL, 
	CONSTRAINT registered_model_alias_pk PRIMARY KEY (name, alias), 
	CONSTRAINT registered_model_alias_name_fkey FOREIGN KEY(name) REFERENCES registered_models (name) ON DELETE cascade ON UPDATE cascade
);
CREATE TABLE datasets (
	dataset_uuid VARCHAR(36) NOT NULL, 
	experiment_id INTEGER NOT NULL, 
	name VARCHAR(500) NOT NULL, 
	digest VARCHAR(36) NOT NULL, 
	dataset_source_type VARCHAR(36) NOT NULL, 
	dataset_source TEXT NOT NULL, 
	dataset_schema TEXT, 
	dataset_profile TEXT, 
	CONSTRAINT dataset_pk PRIMARY KEY (experiment_id, name, digest), 
	FOREIGN KEY(experiment_id) REFERENCES experiments (experiment_id)
);
CREATE TABLE inputs (
	input_uuid VARCHAR(36) NOT NULL, 
	source_type VARCHAR(36) NOT NULL, 
	source_id VARCHAR(36) NOT NULL, 
	destination_type VARCHAR(36) NOT NULL, 
	destination_id VARCHAR(36) NOT NULL, 
	CONSTRAINT inputs_pk PRIMARY KEY (source_type, source_id, destination_type, destination_id)
);
CREATE TABLE input_tags (
	input_uuid VARCHAR(36) NOT NULL, 
	name VARCHAR(255) NOT NULL, 
	value VARCHAR(500) NOT NULL, 
	CONSTRAINT input_tags_pk PRIMARY KEY (input_uuid, name)
);
CREATE INDEX index_datasets_dataset_uuid ON datasets (dataset_uuid);
CREATE INDEX index_datasets_experiment_id_dataset_source_type ON datasets (experiment_id, dataset_source_type);
CREATE INDEX index_inputs_input_uuid ON inputs (input_uuid);
CREATE INDEX index_inputs_destination_type_destination_id_source_type ON inputs (destination_type, destination_id, source_type);
CREATE TABLE `schema_version` (`version` text NOT NULL,PRIMARY KEY (`version`));
CREATE TABLE `dashboards` (`id` uuid,`created_at` datetime,`updated_at` datetime,`is_archived` numeric,`name` text,`description` text,`app_id` uuid,PRIMARY KEY (`id`),CONSTRAINT `fk_dashboards_app` FOREIGN KEY (`app_id`) REFERENCES `apps`(`id`));
CREATE TABLE IF NOT EXISTS "tags" ("key" VARCHAR(250) NOT NULL,value VARCHAR(5000),run_uuid VARCHAR(32) NOT NULL,CONSTRAINT tag_pk PRIMARY KEY ("key", run_uuid),FOREIGN KEY(run_uuid) REFERENCES runs (run_uuid),CONSTRAINT `fk_runs_tags` FOREIGN KEY (`run_uuid`) REFERENCES `runs`(`run_uuid`) ON DELETE CASCADE);
CREATE TABLE IF NOT EXISTS "experiment_tags" ("key" VARCHAR(250) NOT NULL,value VARCHAR(5000),experiment_id INTEGER NOT NULL,CONSTRAINT experiment_tag_pk PRIMARY KEY ("key", experiment_id),FOREIGN KEY(experiment_id) REFERENCES experiments (experiment_id),CONSTRAINT `fk_experiments_tags` FOREIGN KEY (`experiment_id`) REFERENCES `experiments`(`experiment_id`) ON DELETE CASCADE);
CREATE TABLE IF NOT EXISTS "runs" (run_uuid VARCHAR(32) NOT NULL,name VARCHAR(250),source_type VARCHAR(20),source_name VARCHAR(500),entry_point_name VARCHAR(50),user_id VARCHAR(256),status VARCHAR(9),start_time BIGINT,end_time BIGINT,source_version VARCHAR(50),lifecycle_stage VARCHAR(20),artifact_uri VARCHAR(200),experiment_id INTEGER,deleted_time BIGINT,`row_num` bigint,CONSTRAINT run_pk PRIMARY KEY (run_uuid),CONSTRAINT runs_lifecycle_stage CHECK (lifecycle_stage IN ('active', 'deleted')),CONSTRAINT source_type CHECK (source_type IN ('NOTEBOOK', 'JOB', 'LOCAL', 'UNKNOWN', 'PROJECT')),FOREIGN KEY(experiment_id) REFERENCES experiments (experiment_id),CHECK (status IN ('SCHEDULED', 'FAILED', 'FINISHED', 'RUNNING', 'KILLED')),CONSTRAINT `fk_experiments_runs` FOREIGN KEY (`experiment_id`) REFERENCES `experiments`(`experiment_id`) ON DELETE CASCADE);
CREATE TABLE `namespaces` (`id` integer,`code` text NOT NULL,`description` text,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`default_experiment_id` integer NOT NULL,PRIMARY KEY (`id`),CONSTRAINT `uni_namespaces_code` UNIQUE (`code`));
CREATE INDEX `idx_namespaces_deleted_at` ON `namespaces`(`deleted_at`);
CREATE INDEX `idx_namespaces_code` ON `namespaces`(`code`);
CREATE TABLE IF NOT EXISTS "experiments" (experiment_id INTEGER NOT NULL,`name` varchar(256) NOT NULL,artifact_location VARCHAR(256),lifecycle_stage VARCHAR(32),creation_time BIGINT,last_update_time BIGINT,`namespace_id` integer NOT NULL,CONSTRAINT experiment_pk PRIMARY KEY (experiment_id),CONSTRAINT experiments_lifecycle_stage CHECK (lifecycle_stage IN ('active', 'deleted')),UNIQUE (name),CONSTRAINT `fk_namespaces_experiments` FOREIGN KEY (`namespace_id`) REFERENCES `namespaces`(`id`) ON DELETE CASCADE);
CREATE TABLE IF NOT EXISTS "apps" (`id` uuid,`created_at` datetime,`updated_at` datetime,`is_archived` numeric,`type` text NOT NULL,`state` text,`namespace_id` integer NOT NULL,PRIMARY KEY (`id`),CONSTRAINT `fk_namespaces_apps` FOREIGN KEY (`namespace_id`) REFERENCES `namespaces`(`id`) ON DELETE CASCADE);
CREATE TABLE `contexts` (`id` integer,`json` JSON NOT NULL,PRIMARY KEY (`id`),CONSTRAINT `uni_contexts_json` UNIQUE (`json`));
CREATE INDEX `idx_contexts_json` ON `contexts`(`json`);
CREATE TABLE `metrics` (`key` varchar(250) NOT NULL,`value` double precision NOT NULL,`timestamp` integer NOT NULL,`run_uuid` text NOT NULL,`step` integer NOT NULL DEFAULT 0,`is_nan` numeric NOT NULL DEFAULT false,`iter` integer,`context_id` integer NOT NULL,PRIMARY KEY (`key`,`value`,`timestamp`,`run_uuid`,`step`,`is_nan`,`context_id`),CONSTRAINT `fk_metrics_context` FOREIGN KEY (`context_id`) REFERENCES `contexts`(`id`));
CREATE INDEX `idx_metrics_iter` ON `metrics`(`iter`);
CREATE INDEX `idx_metrics_run_id` ON `metrics`(`run_uuid`);
CREATE TABLE `latest_metrics` (`key` varchar(250) NOT NULL,`value` double precision NOT NULL,`timestamp` integer,`step` integer NOT NULL,`is_nan` numeric NOT NULL,`run_uuid` text NOT NULL,`last_iter` integer,`context_id` integer NOT NULL,PRIMARY KEY (`key`,`run_uuid`,`context_id`),CONSTRAINT `fk_latest_metrics_context` FOREIGN KEY (`context_id`) REFERENCES `contexts`(`id`));
CREATE INDEX `idx_latest_metrics_run_id` ON `latest_metrics`(`run_uuid`);
CREATE TABLE IF NOT EXISTS "params" ("key" VARCHAR(250) NOT NULL,`value` varchar(8000) NOT NULL,run_uuid VARCHAR(32) NOT NULL,CONSTRAINT param_pk PRIMARY KEY ("key", run_uuid),FOREIGN KEY(run_uuid) REFERENCES runs (run_uuid),CONSTRAINT `fk_runs_params` FOREIGN KEY (`run_uuid`) REFERENCES `runs`(`run_uuid`) ON DELETE CASCADE);
