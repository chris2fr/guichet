# Guichet

[![Build Status](https://drone.resdigita.org/api/badges/Deuxfleurs/guichet/status.svg?ref=refs/heads/main)](https://drone.resdigita.org/Deuxfleurs/guichet)

Guichet is a simple LDAP web interface for the following tasks:

- self-service password change
- administration of the LDAP directory
- inviting new users to create accounts

Guichet works well with the [Bottin](https://bottin.eu) LDAP server.
Currently, Guichet's templates are only in French as it has been created for
the [Deuxfleurs](https://resdigita.org) collective.
We would gladly merge a pull request with an English transaltion !

A Docker image is provided on the [Docker hub](https://hub.docker.com/r/lxpz/guichet_amd64).
An exemple for running Guichet on a Nomad cluster can be found in `guichet.hcl.exemple`.

Guichet takes a single command line argument, `-config <filename>`, which is the
path to its config file (defaults to `./config.json`).
The configuration file is a JSON file whose contents is described in the following section.

Guichet is licensed under the terms of the GPLv3.


## Building Guichet

Guichet requires go 1.13 or later.

To build Guichet, clone this repository outside of your `$GOPATH`.
Then, run `make` in the root of the repo.


## Configuration of Guichet

Guichet is configured using a simple JSON config file which is a dictionnary whose keys
are described below. An exemple is provided in a further section.

### HTTP listen address

- `http_bind_addr`: which IP and port to bind on for the HTTP web interface. Guichet does not support HTTPS, use a reverse proxy for that.

### LDAP server configuration

- `ldap_server_addr`: the address of the LDAP server to use
- `ldap_tls` (boolean): whether to attempt STARTTLS on the LDAP connection
- `base_dn`: the base LDAP object under which we are allowed to view and modify objects

### User and group configuration

- `user_base_dn`: the base LDAP object that contains user accounts
- `user_name_attr`: the search attribute for user identifiers (usually `uid` or `cn`)
- `group_base_dn` and `group_name_attr`: same for groups

### Administration configuration

- `admin_account`: DN of an LDAP account that has administrative privilege. If such an account is set, the admin can log in by entering the full DN in the login form.
- `group_can_admin`: DN of a LDAP group whose members are given administrative privilege

### Invitation configuration

Guichet supports a mechanism where users can send invitations by email to other users.
The ability to send such invitations can be configured to be restricted to a certain group of users.
Invitation codes are created as temporary LDAP objects in a special folder.

- `group_can_invite`: the LDAP DN of a group whose members are allowed to send invitations to new users
- `invitation_base_dn`: the LDAP folder in which invitation codes are stored
- `invitation_name_attr`: just use `cn`
- `invited_mail_format`: automatically set the invited user's email to this string, where `{}` is replaced by the created username (ex: `{}@resdigita.org`)
- `invited_auto_groups` (list of strings): a list of DNs of LDAP groups

#### Email configuration

Guichet can send an invitation link by email. To do so, an SMTP server must be configured:

- `smtp_server`: the host and port of the mail server
- `smtp_username` and `smtp_password`: the username and password to log in with on the mail server
- `mail_from`: the sender email address for the invitation message
- `web_address`: the base web address of the Guichet service (used for building the invitation link)

## exemple configuration

This is a subset of the configuration we use on Deuxfleurs:

```
{
  "http_bind_addr": ":9991",
  "ldap_server_addr": "ldap://bottin2.service.2.cluster.resdigita.org:389",

  "base_dn": "dc=deuxfleurs,dc=fr",
  "user_base_dn": "ou=users,dc=deuxfleurs,dc=fr",
  "user_name_attr": "cn",
  "group_base_dn": "ou=groups,dc=deuxfleurs,dc=fr",
  "group_name_attr": "cn",

  "admin_account": "cn=admin,dc=deuxfleurs,dc=fr",
  "group_can_admin": "cn=admin,ou=groups,dc=deuxfleurs,dc=fr",
  "group_can_invite": "cn=asso_deuxfleurs,ou=groups,dc=deuxfleurs,dc=fr"
}
```

Here is an exemple of Bottin ACLs that may be used to support Guichet invitations:

```
  "acl": [
		"*,dc=deuxfleurs,dc=fr::read:*:* !userpassword",
		"*::read modify:SELF:*",
		"ANONYMOUS::bind:*,ou=users,dc=deuxfleurs,dc=fr:",
		"ANONYMOUS::bind:cn=admin,dc=deuxfleurs,dc=fr:",
		"*,ou=services,ou=users,dc=deuxfleurs,dc=fr::bind:*,ou=users,dc=deuxfleurs,dc=fr:*",
		"*,ou=services,ou=users,dc=deuxfleurs,dc=fr::read:*:*",

		"*:cn=asso_deuxfleurs,ou=groups,dc=deuxfleurs,dc=fr:add:*,ou=invitations,dc=deuxfleurs,dc=fr:*",
		"ANONYMOUS::bind:*,ou=invitations,dc=deuxfleurs,dc=fr:",
		"*,ou=invitations,dc=deuxfleurs,dc=fr::delete:SELF:*",

		"*:cn=asso_deuxfleurs,ou=groups,dc=deuxfleurs,dc=fr:add:*,ou=users,dc=deuxfleurs,dc=fr:*",
		"*,ou=invitations,dc=deuxfleurs,dc=fr::add:*,ou=users,dc=deuxfleurs,dc=fr:*",

		"*:cn=asso_deuxfleurs,ou=groups,dc=deuxfleurs,dc=fr:modifyAdd:cn=email,ou=groups,dc=deuxfleurs,dc=fr:*",
		"*,ou=invitations,dc=deuxfleurs,dc=fr::modifyAdd:cn=email,ou=groups,dc=deuxfleurs,dc=fr:*",
		"*:cn=asso_deuxfleurs,ou=groups,dc=deuxfleurs,dc=fr:modifyAdd:cn=seafile,ou=groups,dc=deuxfleurs,dc=fr:*",
		"*,ou=invitations,dc=deuxfleurs,dc=fr::modifyAdd:cn=seafile,ou=groups,dc=deuxfleurs,dc=fr:*",

		"cn=admin,dc=deuxfleurs,dc=fr::read add modify delete:*:*",
		"*:cn=admin,ou=groups,dc=deuxfleurs,dc=fr:read add modify delete:*:*"
  ]
```

Consult [this directory](https://git.resdigita.org/Deuxfleurs/infrastructure/src/branch/main/app/directory/config)
to view the full configuration in use on Deuxfleurs.
