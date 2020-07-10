# OCRmyPDF-watchdog

Is the combination of two projects that both seem not to yet work as intended on my Synology NAS.
The Synology NAS provides a `scans/user` directory that gets the scanned PDFs.
All users currently get their own container that does the optical image recognition.
The user is supposed to be able to see his OCR'ed documents in their respective `/homes/user/Drive/Documents` directory.
The problems that need to be tackled are the ownership and usergroups that those resulting files get when put into those Synology Drive folders.

## Inspirating projects

- reactive, kind of interrupt based approach in Node.js: https://github.com/pedropombeiro/OCRmyPDF-watchdog/tree/master-watchdog  
- proactive, polling approach written in Go: https://github.com/bernmic/ocrmypdf-watchdog  

The ractive approach should in theory use less CPU and should only get going when the scan directories actually contain a new legitimate PDF file.

## Setup

Modify the `docker-compose.yml` file to have your needed volumes mounted to the correct host directory and `PUID`(UID of the output file) as well as the `PGID` (GID of the output file) set to the owner of the `/homes/user/Drive/Documents` folder.

Afterwards you simply need to run inside this directory:

```shell
make
```

This creates a few intermediate images that should be cleaned up.
You can either do this manually or with `make clean`.
*BUT* be aware of the possible consequences that might unintentionally occur if you execute this.
You will be *WARNED* before continuing with the cleanup.

## What currently works

*Write* and *Create* events trigger the new file in the `in` directory to be checked for their file type.
Only proper PDF files are actually OCR'ed and then put into the target `out` directory.

## TODO

- Needs to be tested on the actual Synology Diskstation target.
- Check if permissions are correct
- Support either multiple volumes or one valume with individual subfolders instead of having to have one container per user(kinda hefty with its 500MB).
