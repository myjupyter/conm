# Connect through an SSH tunnel

Databases often live in a private network that only a bastion (jump) host can reach. A
transport lets a connection go through that bastion: conm opens an SSH tunnel when the
connection is used, sends the client through it, and closes it when nothing uses it any more.
You never run `ssh -L` by hand.

It takes three pieces:
1. an **SSH connection** to the bastion: where it is and how to log in;
2. a **transport** that tunnels through that SSH connection;
3. the **database connection** set to use the transport.

One SSH connection can back several transports, and one transport can serve any number of
connections, of any database type.

## Without a transport

The database below is at `postgres:5432`, an address only the bastion can resolve. From your
machine the connection fails:

```sh
conm postgres ping orders
```

![A connection to a private host fails without a transport](../demo/transport-direct.gif)

## 1. Add the SSH connection

Add the bastion like any other connection: host, port, user, and how to log in. Auth can be
`agent` (your SSH agent), `key` (a private key file, or a key kept in the form or the system
keyring) or `password`. Test it before saving, and give it a name you will recognise, here
`bastion`.

```sh
conm ssh add
```

![Adding the bastion as an SSH connection](../demo/transport-ssh.gif)

The same connection also opens a shell on the bastion: `conm ssh bastion`.

## 2. Create the transport

Open the transports screen from the connections table and add a transport. Give it a name and
pick the SSH connection it goes through in `via`. The rest is optional:

- `local port`: the port of the tunnel on your machine; a free one is picked when empty;
- `connect timeout`: how long to wait for the tunnel, 10 seconds by default;
- `keepalive`: seconds between probes that keep an idle tunnel open;
- `description`: free text.

```sh
conm
```

![Creating a transport through the bastion](../demo/transport-create.gif)

The transports screen shows which connections use each transport.

## 3. Use the transport

Edit the database connection (or set it while adding one) and, in its transport section, set
the type to `ssh` and pick the transport. Keep `host` and `port` as the bastion sees them, here
`postgres:5432`, not `localhost`.

```sh
conm postgres edit orders
```

![Using the transport: ping, a psql session and a ping from the shell](../demo/transport-use.gif)

That's it. Ping, sessions and `exec` now go through the tunnel, from the table and from the
shell alike:

```sh
conm postgres ping orders
conm postgres orders
conm postgres exec orders -- "select 1;"
```
