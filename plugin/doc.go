/*
Package plugin is responsible for everything related to downloading plugins and parsing them for specific data. It can
load plugins for different sources and determine whether they are up-to-date.

# Details
The entire process of loading plugins & verifying they are up-to-date is done as follows:

First of all, the package will try to find the correct provider for each plugin. The first provider that can
successfully identify the plugin will be chosen as the provider. This means that more 'complex' providers should be
checked first. If no match is found, the process stops, as not all plugins could be identified.

When all the plugins were found, every plugin returns the latest available version. These are checked to see if they
differ from the previously used version. If this is the case for a plugin, the plugin's provider will execute the
necessary steps to ensure we have the latest version downloaded. If there are any changes compared to the previous
version the server should be recompiled.
*/
package plugin
