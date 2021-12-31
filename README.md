# Yet Another GUI for [annie](https://github.com/iawia002/annie)

- Took inspiration from [fanaticscripter/annie-mingui](https://github.com/fanaticscripter/annie-mingui)
- [GPLv3](COPYING)

# Installation

### Linux

- Download from [GitHub Actions](https://github.com/135e2/annie-gtk/actions) or [Release](https://github.com/135e2/annie-gtk/releases)

- Unzip it (make sure that annie-gtk.ui and logo.png is in the directory)

- Install [FFmpeg](https://www.ffmpeg.org) to merge videos
  
  > **Note**: FFmpeg does not affect the download, only affects the final file merge.

```bash
# Install via apt-get (Ubuntu/Debian)
sudo apt-get install ffmpeg
```

- Execute *annie-gtk*

### Windows

- Download from [GitHub Actions](https://github.com/135e2/annie-gtk/actions) or [Release](https://github.com/135e2/annie-gtk/releases)

- Unzip it (make sure that annie-gtk.ui and logo.png is in the directory)

- Install [FFmpeg](https://www.ffmpeg.org) to merge videos
  
  > **Note**: FFmpeg does not affect the download, only affects the final file merge.

```powershell
# Install via scoop (Windows)
scoop install ffmpeg
```

- Install *gtk3-runtime-3.24.29-2021-04-29-ts-win64.exe*

- Execute *annie-gtk.exe*