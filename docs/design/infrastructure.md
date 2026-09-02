# Infrastructure (reference deployment)

## Virtual infrastructure
Example private network (VPC): `10.0.1.0/24`

### Instances
Type: Ubuntu

Pre-installed:
- Admin user (set your own credentials — do **not** hardcode)
- SSH service

## Provisioning
```
ssh -i user_rsa admin@10.0.1.128
Welcome to Ubuntu (GNU/Linux)
 * Documentation:  https://help.ubuntu.com/
admin@host:~$
```

## References
- Creating a private network in VMware / VirtualBox.

> The original notes contained a lab username/password and hostname; those have
> been removed. Never commit real credentials — use a secrets manager or env vars.
