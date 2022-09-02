# Plugin installation
Unlike a lot of other servers, plugins in saddle are not installed by dragging and dropping a plugin file into the
plugins folder, but rather by adding it in the `saddle.toml` file. The launcher will automatically download the version
specified, and even keep it up to date if applicable.

There are currently two different ways and places to install a plugin from.

### Installing from GitHub
To add a plugin from GitHub (or any remote git repository), add the following entry to your `saddle.toml`:

```toml
[[plugin]]
# The link to the plugin repo, without the 'https://' in front of it. The project in this repository should be a valid
# go modules project.
module = "github.com/author/repository"
# The version of the plugin, any git ref accepted by the go get command, such as a tag, branch name, 'latest' or even
# just the commit hash. In cases like 'latest' and a branch name, the plugin will be automatically updated if an update
# is available. Note that a version starts with 'v', for example 'v1.0.0'.
version = "latest"
```

Of course, replace the module and version to your actual module and the version you want. Saddle will handle everything 
for you from there on, all you need to do is run the launcher again!

### Installing a local plugin
If your plugin is on your local disk, you may find it easier to directly install it from there. This can be done by 
adding the following to your `saddle.toml`:

```toml
[[plugin]]
# May be any absolute or relative path to a valid go modules project. If any changes are detected in the folder since 
# the last time the launcher ran, the server will be recompiled with this new version.
local = "path/to/plugin"
```
