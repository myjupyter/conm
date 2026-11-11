# conm user guide

The pages read in order: each one builds on the one before and links to the next. The examples
use Postgres; every other database works the same way under its own name.

The pages say what to do, not which key does it. Every screen lists its keys: press `?` for the
full list of the screen you are on, and `?` again to close it. Keys can be changed in the
[settings](additional-screens.md#overriding-keybindings).

![Opening the key list in the connections table and in a form](../demo/help.gif)

## Basics

1. [First run](first-run.md): what `conm` shows before you have any connections
2. [Add a connection](add-connection.md): fill in a form, or prefill it from a command line or
   URL you already use
   - [Interactive mode](add-connection.md#interactive-mode)
   - [Prefill from a command line](add-connection.md#prefill-from-a-command-line)
3. [Edit or remove a connection](edit-connection.md): change a saved connection, or delete it
   - [From the connections table](edit-connection.md#from-the-connections-table)
   - [From the shell](edit-connection.md#from-the-shell)
   - [Remove a connection](edit-connection.md#remove-a-connection)
4. [Connect and run queries](connect-and-execute.md): open the database client, or run a query
   or a script
   - [From the connections table](connect-and-execute.md#from-the-connections-table)
   - [From the shell](connect-and-execute.md#from-the-shell)
   - [Run a query or a script from the shell](connect-and-execute.md#run-a-query-or-a-script-from-the-shell)

## Going further

5. [Passwords and secrets](secrets.md): keep passwords in the system keyring instead of
   plain text
   - [The secrets screen](secrets.md#the-secrets-screen)
   - [Use a secret in a connection](secrets.md#use-a-secret-in-a-connection)
   - [Secrets and workspaces](secrets.md#secrets-and-workspaces)
6. [Connect through an SSH tunnel](ssh-transport.md): reach a database behind a bastion host
   - [Without a transport](ssh-transport.md#without-a-transport)
   - [1. Add the SSH connection](ssh-transport.md#1-add-the-ssh-connection)
   - [2. Create the transport](ssh-transport.md#2-create-the-transport)
   - [3. Use the transport](ssh-transport.md#3-use-the-transport)
7. [Workspaces](workspaces.md): keep separate sets of connections and secrets apart
   - [What belongs to a workspace](workspaces.md#what-belongs-to-a-workspace)
   - [From the workspaces screen](workspaces.md#from-the-workspaces-screen)
   - [From the shell](workspaces.md#from-the-shell)

## Customising

8. [Additional screens](additional-screens.md): settings and the info screen
   - [Switching clients](additional-screens.md#switching-clients)
   - [Overriding keybindings](additional-screens.md#overriding-keybindings)
   - [Info screen](additional-screens.md#info-screen)

---

[Start reading: First run →](first-run.md)
