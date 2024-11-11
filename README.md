# coolAD
### Remote AWS Active Directory Environment

#### Prerequisites
- UNIX/LINUX
- AWS region specified in the `AWS_REGION`, `AWS_DEFAULT_REGION` [environment vars](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html#:~:text=command%20line%20parameter.-,AWS_REGION,-The%20AWS%20SDK), or the [config file](https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/setup-credentials.html#setup-credentials-setting-region) @  `~/.aws/config` 
- Go installed if not using `coolAD` binary for Apple M1
- Doormat CLI to authenticate to AWS
- (optional) openldap to test--can install via:
```
Ubuntu
sudo apt-get update
sudo apt-get install ldap-utils

On CentOS/RHEL:
sudo yum install openldap-clients

On macOS:
brew install openldap

```
  

#### Quick Setup Via Binary
- binary compiled on Apple M1
- owners of machine with same architecture just download `cool-ad` and run `./coolAD` (may need chmod +x)
- if mac throws a security message, go to System Settings -> Privacy & Security -> Open Anyway

#### Setup Without Binary
- clone repo; cd into 
- run `go run .`

#### Usage
- first time use, enter `no/n` at cleanup prompt
- checks first for AWS region, exits if not set
- wait about 5 min for resources to be created
- once finished, will output ssh command to openssh to windows server/AD domain controller
- will also output ldapsearch test command; openldap not installed on server
- outputs powershell commands to view AD details, which can be executed on server
- if error recieved when attempting powershell commands, need to wait another few minutes for AD bootstrap completion
- when finished, run again and enter `yes/y` to run cleanup
- cleanup takes about 3-5 minutes

#### Server Info
- ldaps is enabled
- domain controller
- username is `Administrator`
- password is `admin`
- AD domain is `vaultest.com`
- password complexities have been disabled
- once ssh connection established, all commands are run in batch, just type `powershell` to switch
- `exit` to close connection
- Windows EC2 instances employ a "launch" agent for customizing startup parameters
- this instance is using EC2Launch v2 agent which allows startup tasks to be defined via yaml in user-data
- more [info](https://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/ec2launch-v2-settings.html#ec2launch-v2-schema-user-data) on EC2Launch V2 syntax/task defninitions

#### Connection Commands
Example SSH connection string that will be output.  Ensure the key.pem file output has 400 perms:
`ssh -i key.pem -o StrictHostKeyChecking=no Administrator@ec2-34-221-77-219.us-west-2.compute.amazonaws.com`


ldapsearch example cmd can be run from a remote linux box:
`ldapsearch -x -H ldap://ec2-34-221-77-219.us-west-2.compute.amazonaws.com:389 -D "cn=Administrator,cn=Users,dc=vaultest,dc=com" -w admin -b "dc=vaultest,dc=com" -s sub "(objectclass=*)"`


Can also run the following on server to see AD details (if error, will need to wait a bit longer):

> powershell
> Get-ADForest
> Get-ADUser -Filter *
> Get-ADUser -Identity Administrator -Properties *