# Mounts

Mounts allow you to attach directories from the host machine directly into your server containers. This is extremely useful for sharing common resources like maps, plugins, mods, or configurations across multiple servers without duplicating files.

## Creating a Mount

Mounts are created and managed by administrators in the **Admin -> Mounts** interface.

To create a mount, you'll need to specify:

* **Name**: A descriptive name for the mount (e.g., `Shared Maps`).
* **Description**: An optional explanation of what the mount is for.
* **Source Path**: The absolute path on the host machine where the files reside (e.g., `/mnt/storage/shared_maps`).
* **Target Path**: The absolute path inside the server container where the mount will appear (e.g., `/home/container/shared_maps`). Do not use `/home/container` directly, as it will overwrite everything in the container's root.
* **Read Only**: If enabled, the server will not be able to modify, delete, or create files in the mounted directory.
* **User Mountable**: If enabled, server owners with the appropriate permissions can choose whether or not to attach this mount to their servers from the server's Mounts page.
* **Navigable**: If enabled, the mount will be visibly injected into the server's File Manager and SFTP, allowing users to browse its contents directly from the panel.

## Attaching Mounts to Servers

### Automatic Attachment
If a mount is assigned directly to a Node and a Server is assigned to that Node, the mount will be automatically attached to that Server unless it is designated as "User Mountable". 
Additionally, Packages can also specify Mounts to be automatically attached when a server is created using that Package.

### User Mountable Status
When you check the **User Mountable** option during mount creation, server owners are given the choice of whether or not to attach that mount to their server. 

Server owners can manage these mounts by navigating to their Server Console and selecting **Mounts** from the side navigation. From there, they can click "Attach to Server" or "Detach from Server".

> [!IMPORTANT]
> Because mount configurations are sent to the Axis daemon during the server startup sequence, **you must start or restart your server** for any newly attached or detached mounts to take effect. If you modify a mount, you must also restart your server before those changes apply to the running container. ("Please do" - says me after going crazy about why they were not appearing)

## Navigable Mounts

The Virtual File System (VFS) in Birdactyl allows you to expose mounts securely to your users. When you check the **Navigable** option on a mount, the Axis daemon dynamically binds the mount into the user's File Manager.

For example, if you set the target to `/home/container/textures` and mark it as Navigable, an empty directory called `textures` will be displayed at the root of the server's File Manager. Clicking on it will transparently redirect the user's view into the host system mount.

If a mount is *not* marked as Navigable, its files will still be accessible to the server process itself (e.g., the game engine can load files from it), but users will not be able to interact with its files through the panel's File Manager or SFTP.
