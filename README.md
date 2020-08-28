# ctrun

## Building

You need to have Docker installed and running

```shell
$ make
```

This will create `bin/ctrun`

## Running

`ctrun` uses an S3 bucket as storage, you will need to define and pass auth info about the S3:

- `--endpoint` (`$S3_ENPOINT`)
- `--access-key` (`$ACCESS_KEY_ID`)
- `--secret-key-id` (`$SECRET_KEY_ID`)
- `--bucket` (`$S3_BUCKET`)

Once you define these environment variables you can simply run:

```shell
$ ./bin/ctrun
🚀 Server started
```

Or

```shell
$ ./bin/ctrun --endpoint s3.amazon.com --bucket ctrun --access-key ACCESS_KEY --secret-key-id SECRET_KEY_ID
🚀 Server started
```
