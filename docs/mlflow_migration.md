## Using FastTrackML with an Existing MLFlow Database

### Prerequisites

-   FastTrackML version 0.3.0 or later.
-   An MLFlow database created with MLFlow version 1.21.0 to 2.10.0.

### Compatibility with MLFlow Database

-   **Supported Versions:** FastTrackML is compatible with MLFlow databases created using MLFlow versions 1.21.0 to 2.10.0.

### Database Compatibility

-   When integrating an existing MLFlow database, FastTrackML will transition its schema to align with MLFlow version 1.29.0-2.10.0. Consequently, the database will no longer be compatible with earlier MLFlow versions. If compatibility with prior versions is needed, use a copy of the database instead.
-   FastTrackML may introduce schema alterations in other places that have remained compatible with MLFlow so far. However, complete compatibility cannot be assured in every scenario. In case of uncertainty, it's recommended to use a copy of the database.

### Setting Up FastTrackML

-   To point FastTrackML to an existing database, specify your current database URI using the `--database-uri` parameter or the `FML_DATABASE_URI` environment variable.
-   The default artifact root can be set via the `--artifact-root` parameter or the `FML_ARTIFACT_ROOT` environment variable. Note that this functionality is available only in FastTrackML 0.3.0 or later.
-   FastTrackML currently supports artifact storage on either a filesystem or on S3 or S3-compatible storage platforms.
-   In case of utilizing an S3-compatible storage platform (e.g., Minio), configure the `FML_S3_ENDPOINT_URI` environment variable to correspond with your `MLFLOW_S3_ENDPOINT_URL`.

### Example

#### Postgres
If you initiated an MLFlow tracking server using the subsequent command, you can substitute it with the equivalent FastTrackML command:

```console
# MLFlow  
mlflow server --backend-store-uri postgresql://postgres:postgres@localhost/mlflow --default-artifact-root s3://mlflow  

# FastTrackML  
fml server --database-uri postgresql://postgres:postgres@localhost/mlflow --artifact-root s3://mlflow
```

#### Sqlite

```console
# MLFlow  
mlflow server --backend-store-uri sqlite:///mlflow.db 

# FastTrackML  
fml server --database-uri sqlite://mlflow.db 
```
#### Using environment variable

```console
# MLFlow  
mlflow server --backend-store-uri postgresql://postgres:postgres@localhost/mlflow --default-artifact-root s3://mlflow  

# FastTrackML  
export FML_DATABASE_URI=postgresql://postgres:postgres@localhost/mlflow
export FML_ARTIFACT_ROOT=s3://mlflow
fml server
```