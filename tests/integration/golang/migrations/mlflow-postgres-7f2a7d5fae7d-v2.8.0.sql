create table experiments
(
    experiment_id     serial
        constraint experiment_pk
            primary key,
    name              varchar(256) not null
        unique,
    artifact_location varchar(256),
    lifecycle_stage   varchar(32)
        constraint experiments_lifecycle_stage
            check ((lifecycle_stage)::text = ANY
        ((ARRAY ['active'::character varying, 'deleted'::character varying])::text[])),
    creation_time     bigint,
    last_update_time  bigint
);

alter table experiments
    owner to postgres;

create table runs
(
    run_uuid         varchar(32) not null
        constraint run_pk
            primary key,
    name             varchar(250),
    source_type      varchar(20)
        constraint source_type
            check ((source_type)::text = ANY
        ((ARRAY ['NOTEBOOK'::character varying, 'JOB'::character varying, 'LOCAL'::character varying, 'UNKNOWN'::character varying, 'PROJECT'::character varying])::text[])),
    source_name      varchar(500),
    entry_point_name varchar(50),
    user_id          varchar(256),
    status           varchar(9)
        constraint runs_status_check
            check ((status)::text = ANY
                   ((ARRAY ['SCHEDULED'::character varying, 'FAILED'::character varying, 'FINISHED'::character varying, 'RUNNING'::character varying, 'KILLED'::character varying])::text[])),
    start_time       bigint,
    end_time         bigint,
    source_version   varchar(50),
    lifecycle_stage  varchar(20)
        constraint runs_lifecycle_stage
            check ((lifecycle_stage)::text = ANY
                   ((ARRAY ['active'::character varying, 'deleted'::character varying])::text[])),
    artifact_uri     varchar(200),
    experiment_id    integer
        references experiments,
    deleted_time     bigint
);

alter table runs
    owner to postgres;

create table tags
(
    key      varchar(250) not null,
    value    varchar(5000),
    run_uuid varchar(32)  not null
        references runs,
    constraint tag_pk
        primary key (key, run_uuid)
);

alter table tags
    owner to postgres;

create index index_tags_run_uuid
    on tags (run_uuid);

create table metrics
(
    key       varchar(250)                not null,
    value     double precision            not null,
    timestamp bigint                      not null,
    run_uuid  varchar(32)                 not null
        references runs,
    step      bigint  default '0'::bigint not null,
    is_nan    boolean default false       not null,
    constraint metric_pk
        primary key (key, timestamp, step, run_uuid, value, is_nan)
);

alter table metrics
    owner to postgres;

create index index_metrics_run_uuid
    on metrics (run_uuid);

create table params
(
    key      varchar(250)  not null,
    value    varchar(8000) not null,
    run_uuid varchar(32)   not null
        references runs,
    constraint param_pk
        primary key (key, run_uuid)
);

alter table params
    owner to postgres;

create index index_params_run_uuid
    on params (run_uuid);

create table alembic_version
(
    version_num varchar(32) not null
        constraint alembic_version_pkc
            primary key
);
insert into alembic_version VALUES('7f2a7d5fae7d');

alter table alembic_version
    owner to postgres;

create table experiment_tags
(
    key           varchar(250) not null,
    value         varchar(5000),
    experiment_id integer      not null
        references experiments,
    constraint experiment_tag_pk
        primary key (key, experiment_id)
);

alter table experiment_tags
    owner to postgres;

create table latest_metrics
(
    key       varchar(250)     not null,
    value     double precision not null,
    timestamp bigint,
    step      bigint           not null,
    is_nan    boolean          not null,
    run_uuid  varchar(32)      not null
        references runs,
    constraint latest_metric_pk
        primary key (key, run_uuid)
);

alter table latest_metrics
    owner to postgres;

create index index_latest_metrics_run_uuid
    on latest_metrics (run_uuid);

create table registered_models
(
    name              varchar(256) not null
        constraint registered_model_pk
            primary key,
    creation_time     bigint,
    last_updated_time bigint,
    description       varchar(5000)
);

alter table registered_models
    owner to postgres;

create table model_versions
(
    name              varchar(256) not null
        references registered_models
            on update cascade,
    version           integer      not null,
    creation_time     bigint,
    last_updated_time bigint,
    description       varchar(5000),
    user_id           varchar(256),
    current_stage     varchar(20),
    source            varchar(500),
    run_id            varchar(32),
    status            varchar(20),
    status_message    varchar(500),
    run_link          varchar(500),
    storage_location  varchar(500),
    constraint model_version_pk
        primary key (name, version)
);

alter table model_versions
    owner to postgres;

create table registered_model_tags
(
    key   varchar(250) not null,
    value varchar(5000),
    name  varchar(256) not null
        references registered_models
            on update cascade,
    constraint registered_model_tag_pk
        primary key (key, name)
);

alter table registered_model_tags
    owner to postgres;

create table model_version_tags
(
    key     varchar(250) not null,
    value   varchar(5000),
    name    varchar(256) not null,
    version integer      not null,
    constraint model_version_tag_pk
        primary key (key, name, version),
    foreign key (name, version) references model_versions
        on update cascade
);

alter table model_version_tags
    owner to postgres;

create table registered_model_aliases
(
    alias   varchar(256) not null,
    version integer      not null,
    name    varchar(256) not null
        constraint registered_model_alias_name_fkey
            references registered_models
            on update cascade on delete cascade,
    constraint registered_model_alias_pk
        primary key (name, alias)
);

alter table registered_model_aliases
    owner to postgres;

create table datasets
(
    dataset_uuid        varchar(36)  not null,
    experiment_id       integer      not null
        references experiments,
    name                varchar(500) not null,
    digest              varchar(36)  not null,
    dataset_source_type varchar(36)  not null,
    dataset_source      text         not null,
    dataset_schema      text,
    dataset_profile     text,
    constraint dataset_pk
        primary key (experiment_id, name, digest)
);

alter table datasets
    owner to postgres;

create index index_datasets_dataset_uuid
    on datasets (dataset_uuid);

create index index_datasets_experiment_id_dataset_source_type
    on datasets (experiment_id, dataset_source_type);

create table inputs
(
    input_uuid       varchar(36) not null,
    source_type      varchar(36) not null,
    source_id        varchar(36) not null,
    destination_type varchar(36) not null,
    destination_id   varchar(36) not null,
    constraint inputs_pk
        primary key (source_type, source_id, destination_type, destination_id)
);

alter table inputs
    owner to postgres;

create index index_inputs_destination_type_destination_id_source_type
    on inputs (destination_type, destination_id, source_type);

create index index_inputs_input_uuid
    on inputs (input_uuid);

create table input_tags
(
    input_uuid varchar(36)  not null,
    name       varchar(255) not null,
    value      varchar(500) not null,
    constraint input_tags_pk
        primary key (input_uuid, name)
);

alter table input_tags
    owner to postgres;

