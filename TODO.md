# Todo

[] - Add Windows paths default discoveries - GDrive: G:\My Drive. Dropbox: ~\Dropbox. iCloud Drive for later.
[] - Add Windows build (basically output has got to have .exe, then we can symlink it to normal `dotsync).
    Figure out how to do so with github actions and goreleaser.
[] - In taskfile, customize build command to work on windows (basically add .exe and for date function use `Get-Date -Format "yyyy-MM-dd"`)
[] - Add warning that windows 10-11 needs admin / developer mode on to create symlinks.

Final output desired:
- Windows can install dotsync via install script and also build it locally by itself
- Windows detects most common paths for Google Drive and Dropbox

MAKE NOTE ABOUT FILE STREAMING! (for gdrive, idk about dropbox, need to test)

## Notes

Apparently iCloud drive exists in windows too (need to test / maybe keep for later)
