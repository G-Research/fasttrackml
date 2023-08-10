#!/bin/sh
gem install 'sequel'
gem install 'sqlite3'
gem install 'pg'
sequel -C $INPUT_DATABASE_URI $OUTPUT_DATABASE_URI
