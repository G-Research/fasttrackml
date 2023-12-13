PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "metrics" (
	"key" VARCHAR(250) NOT NULL, 
	value FLOAT NOT NULL, 
	context_id INTEGER NOT NULL,
	context_null_id INTEGER NULL,
	CONSTRAINT metric_pk PRIMARY KEY ("key", value, context_id), 
	FOREIGN KEY(context_id) REFERENCES contexts (id), 
	FOREIGN KEY(context_null_id) REFERENCES contexts (id)
);
CREATE TABLE `contexts` (
        `id` integer,
	`json` JSON NOT NULL UNIQUE,
	PRIMARY KEY (`id`));
CREATE INDEX `idx_contexts_json` ON `contexts`(`json`);
COMMIT;
