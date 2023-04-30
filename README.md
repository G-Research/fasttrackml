# fasttrack
Rewrite of the MLFlow tracking server with a focus on scalability

| ⚠️ This is a work in progress ⚠️ |
| :------------------------------: |
| 🗒️ name subject to change |


### Quickstart

For the full guide, see [docs/quickstart.md](docs/quickstart.md).

FastTrack can be run using the following command:

```bash
docker run --rm -p 5000:5000 -ti gresearch/fasttrack
```

Verify that you can see the UI by navigating to http://localhost:5000/.

![FastTrack UI](docs/images/main_ui.jpg)

For more info, `--help` is your friend!

#### Encryption at rest

To use an encrypted SQLite database, use the query parameter `_key` in the DSN:

```
docker run --rm -ti fasttrack server --database-uri 'sqlite:///data/fasttrack.db?_key=passphrase'
```

### License

Copyright 2022-2023 G-Research

Copyright 2019-2022 Aimhub, Inc.

Copyright 2018 Databricks, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use these files except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
