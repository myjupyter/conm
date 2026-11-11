# Passwords and secrets

Every connection says where its password comes from. That choice is the `provider` field of the
form, and today there are two places a password can live:

| Provider | Where the password is | What the connection file holds |
|---|---|---|
| `literal` | in the connection file itself, as plain text | `password = 'demo-secret'` |
| `keyring` | in the system keyring: Keychain on macOS, Credential Manager on Windows, Secret Service (GNOME Keyring, KWallet) on Linux | `password = 'keyring:orders'` |
| `none` | nowhere: the connection sends no password | `password = 'none:'` |

`literal` is the default and the quickest to set up, but anyone who can read the file can read
the password. Use the keyring for anything you care about: the password never touches conm's
files.

> [!NOTE]
> The system keyring and plain text are the only stores for now. More are planned, such as
> Vault, gopass, 1Password and KeePassXC.

An SSH connection with key auth can point at a private key file, or keep the key itself in the
form (plain text) or in the keyring.

## The secrets screen

Keyring passwords are managed as *secrets* on the secrets screen, opened from the connections
table. A secret has:
- **id**: the name you pick it by;
- **location**: where it lives in the keyring, `account` or `service/account`;
- **password**: written straight to the keyring, never to disk. When editing, leave it empty to
  keep the current one;
- **description**: free text.

```sh
conm
```

![Adding a secret to the keyring](../demo/secrets-add.gif)

The screen shows which connections use each secret. A secret that is still in use cannot be
deleted: conm lists the connections that need it. Deleting an unused secret also removes its
password from the keyring.

## Use a secret in a connection

In the connection's form, set the provider to `keyring` and pick a secret: the secrets screen
opens as a picker, and you can add a new secret from there too. After saving, the connection
file only holds the reference.

```sh
conm postgres edit orders-local
```

![Switching a connection from a plain-text password to a keyring secret](../demo/secrets-use.gif)

## Secrets and workspaces

Every [workspace](workspaces.md) has its own list of secrets, and a secret whose location is a
plain `account` is stored under that workspace's own keyring ID. Such a secret can only be used
from its workspace, and deleting the workspace removes its password from the keyring.

A location with a service, `service/account`, points at that exact keyring entry instead. Two
workspaces that both name `conm/orders` read and write the same password. Leave the service out
unless you want that.

---

[← Connect and run queries](connect-and-execute.md) · [Contents](README.md) · [Connect through an SSH tunnel →](ssh-transport.md)
