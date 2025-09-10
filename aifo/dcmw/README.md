# DICOM Warehouse (DCMW)

Welcome to the DICOM Warehouse (DCMW) project. This tool is designed to manage and process DICOM files efficiently using Go. It provides a robust solution for storing, querying, and retrieving medical imaging data.

## Compilation

To compile the DCMW project, use the following command:

```sh
go build -o bin/dcmw ./cmd/dcmw
```

## Running the Application

After compiling, you can run the application with the following command:

```sh
./bin/dcmw -dirs /folder/to/dicom_data -db_name test.sqlite
```

This command will process the DICOM files located in the specified directory and store the metadata in the specified SQLite database.

## Modifying the database fields

To modify the database fields, follow these steps:

1. Edit the file `internal/models/models.go` to include your desired changes.
2. Navigate to the `internal/processor` folder and run the following command to regenerate the mappings, otherwise the new fields will not be populated:

   ```sh
   go generate
   ```

3. Format the generated code by running:

   ```sh
   go fmt ./...
   ```
