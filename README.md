# Saddle Launcher
The Saddle Launcher is a launcher for the saddle server. It provides a simple way to add and remove plugins, and will
automatically compile the server for you. The launcher can also automatically download new plugin updates when 
available. For more information about saddle itself, please check the 
[saddlemc/saddle](https://github.com/saddlemc/saddle) repository. This repository only contains the launcher, and not 
the server itself.

There are currently no prebuilt launcher binaries available.

### Useful resources
* [How to use](#how-to-use)
* [Plugin installation guide](PLUGINS.md)

## How to use
Before being able to use the launcher, you will need to install the latest version of the 
[Go programming language](https://go.dev/dl) to your computer. This is used by the launcher to bundle in plugins with
the server, and build them into one executable.

With Go installed, put the launcher executable in the folder you want your server to be in. Now, you can open the
launcher. You will notice that the launcher is building your server. The first time you do this, this might take a 
minute depending on your hardware. After that, building should be considerably faster.

When your server has been built, a `server` or `server.exe` (depending on your platform) will appear next to the 
launcher executable. This is your server bundled with all the plugins you might have installed. You should also see a
`saddle.toml` file appear next to the launcher executable. This file can be modified to change build settings. These are 
separate from the server's `config.toml` that will also be generated. At this point, your server will be up & running.
You can close it again by pressing CRTL+C in the server's terminal.
