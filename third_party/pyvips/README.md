# How to create the patch file

If you need to modify the pyvips build script, you can create a patch file to apply to the original file.
Find the original file in the `MODULE.bazel` file, untar and create a patch file.

```shell
diff -u pyvips-2.2.3/pyvips/pyvips_build.py third_party/pyvips/pyvips_build.py > third_party/pyvips/pyvips_build.patch
```

Modify the first two lines of the patch file to point to this:

```shell
--- pyvips/pyvips_build.py	2024-04-28 13:15:50
+++ pyvips/pyvips_build.py	2025-03-02 16:34:34
```
