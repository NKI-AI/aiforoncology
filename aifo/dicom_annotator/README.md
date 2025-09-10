# DICOM Annotator tool

This webbased platform allows to annotate sequences names.

## Development

To make sure the database models in the dicom warehouse and the annotator match, you can use `sqlacodegen`

```shell
sqlacodegen sqlite:///path/to/your/database.db --outfile src/python/research/dicom_annotator/models.py
```

Current version on pip does not support python 3.11. Install the pre-release

```shell
pip install git+https://github.com/agronholm/sqlacodegen.git#egg=sqlacodegen
```

Do not forget to run the linters afterwards using pre-commit.
