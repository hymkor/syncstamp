# syncstamp

`syncstamp` synchronizes file timestamps between two directories when the file contents are identical.

This tool is intended for release engineering and delivery workflows where file timestamps are used as an indicator of modifications.

When preparing a new release package, some files may receive newer timestamps even though their contents are completely unchanged.

`syncstamp` restores the timestamps of such files from the previous release so that unchanged files do not appear modified.

## Features

- Compare files by checksum, not by timestamp
- Synchronize timestamps only when file contents are identical
- Match files by filename only (directory structure is ignored)
- Useful for workflows where files are flattened into a single directory before packaging
- Can either:
  - directly update timestamps
  - or generate shell commands (`touch -r`) for later execution

## Usage

```text
syncstamp [OPTIONS] <SRC-DIR> <DST-DIR>
```

`DST-DIR` file timestamps are synchronized to match files in `SRC-DIR`.

## Options

### `-batch`

Generate shell commands to stdout instead of modifying files directly.

The generated commands use `touch -r`.

Example:

```sh
syncstamp -batch release-v1.0 release-v1.1 > sync.sh
sh sync.sh
```

### `-update`

Update timestamps immediately.

Example:

```sh
syncstamp -update release-v1.0 release-v1.1
```

## File Matching Rules

### Filename-based matching

Files are matched using only the filename portion.

Directory structures are ignored.

Example:

```text
SRC-DIR/lib/foo.dll
DST-DIR/package/foo.dll
```

These files are considered a match because both filenames are `foo.dll`.

This behavior is intentional because packaging workflows often flatten files into a single directory before distribution.

### Content verification

Even when filenames match, timestamps are synchronized only if the file contents are identical.

Content equality is verified using checksums.

### Multiple matching files

When multiple files share the same filename, files are grouped by:

- filename
- checksum

All matching destination files are synchronized.

If multiple source files belong to the same group, the oldest timestamp among the source files is used.

## Typical Workflow

Example:

```text
release-v1.0/
  app.exe
  foo.dll
  bar.dll

release-v1.1/
  app.exe
  foo.dll
  bar.dll
```

Suppose:

- `foo.dll` was not changed between releases
- `bar.dll` was rebuilt or modified
- packaging or copying changed timestamps for all files

Run:

```sh
syncstamp -update release-v1.0 release-v1.1
```

Result:

- `foo.dll` timestamp is restored to the older timestamp from `release-v1.0`
- `bar.dll` keeps its newer timestamp because the contents differ

This makes it easier to identify which files were actually modified between releases.

## Notes

- Files with different contents are ignored
- Directory timestamps are not handled
- Files are matched by basename only

## License

MIT License
