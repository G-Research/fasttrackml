
## Using FastTrackML with an Existing MLFlow Database

### Requirements

-   FastTrackML version 0.3.0 or later.
-   An MLFlow database created with MLFlow 1.21.0 to 2.2.2.

### MLFlow Database Compatibility

-   **Supported Versions:** FastTrackML supports MLFlow databases created with MLFlow 1.21.0 to 2.2.2.
-   **Higher Versions:** Higher versions of MLFlow will require the use of FastTrackML 0.3.0 or later (including support for MLFlow versions 2.3.0 to 2.6.0).

### Database Compatibility Notice

-   When using an existing MLFlow database, FastTrackML will upgrade its schema to the one from MLFlow 1.29.0-2.2.2. As a result, your database will become unusable with lower versions of MLFlow. If compatibility with lower versions is needed, please use a copy of the database instead.
-   FastTrackML may also alter the schema in other places that have remained compatible with MLFlow so far. However, compatibility cannot be guaranteed in all cases. If in doubt, it's recommended to use a copy of the database.

### Configuring FastTrackML

-   To point FastTrackML to an existing database, pass your existing Postgres database URI via the `--database-uri` parameter or via the `FML_DATABASE_URI` environment variable.
-   The default artifact root can be set via the `--artifact-root` parameter or the `FML_ARTIFACT_ROOT` environment variable. Note that this is only supported in FastTrackML 0.3.0 or in nightly releases, usable via the `gresearch/fasttrackml:edge` Docker image.
-   FastTrackML currently supports artifacts stored on a filesystem or on S3 or S3-compatible storage.
-   If you use an S3-compatible storage platform (like Minio), set the `FML_S3_ENDPOINT_URI` environment variable to match your `MLFLOW_S3_ENDPOINT_URL`.

### Example

Here is a concrete example. If you used the following command to start an MLFlow tracking server, you can replace it with the equivalent FastTrackML command:

```console
# MLFlow  
mlflow server --backend-store-uri postgresql://postgres:postgres@localhost/mlflow --default-artifact-root s3://mlflow  
# FastTrackML  
fml server --database-uri postgresql://postgres:postgres@localhost/mlflow --artifact-root s3://mlflow
```
