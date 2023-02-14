# fasttrack
Rewrite of the MLFlow tracking server with a focus on scalability

| ⚠️ This is a work in progress ⚠️ |
| :------------------------------: |
| 🗒️ name subject to change |

### Building

```
docker build -t fasttrack .
```

### Running

```
docker run --rm -ti fasttrack
```

For more info, `--help` is your friend!

#### Encryption at rest

To use an encrypted SQLite database, use the query parameter `_key` in the DSN:

```
docker run --rm -ti fasttrack server --database-uri 'sqlite:///data/fasttrack.db?_key=passphrase'
```

### License

Copyright 2022-2023 G-Research
Copyright 2018 Databricks, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use these files except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
