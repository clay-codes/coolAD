package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"os/exec"

	"github.com/clay-codes/aws-ldap/cloud"
)

var runCleanup bool

func init() {

	// prompt user if they want to run cleanup
	fmt.Print("Would you like to run cleanup? ")
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	runCleanup = strings.ToLower(response) == "yes" || strings.ToLower(response) == "y"

	// authenticate with AWS
	cloud.CheckAuth()

	// creating a session
	cloud.SetRegion()
	if err := cloud.CreateSession(); err != nil {
		log.Fatal(err)
	}

	// creating needed services from session
	if err := cloud.GetSession().CreateServices(); err != nil {
		log.Fatal(err)
	}
}
// appends aws resources to a file in bootstrap
func appendToFile(filename, text string) error {
    file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("failed to open file: %v", err)
    }
    defer file.Close()

    if _, err := file.WriteString(text); err != nil {
        return fmt.Errorf("failed to write to file: %v", err)
    }

    return nil
}
// build environment
func bootStrap() {
	key, err := cloud.CreateKP()
	if err != nil {
		log.Fatal(err)
	}
    err = appendToFile("aws-resources.txt", fmt.Sprintf("key pair             %s\n", key))
    if err != nil {
        log.Fatal(err)
    }
	fmt.Printf("\nkey pair created            %s", key)

	sgid, err := cloud.CreateSG()
	if err != nil {
		log.Fatal(err)
	}
	err = appendToFile("aws-resources.txt", fmt.Sprintf("security group       %s\n", sgid))
    if err != nil {
        log.Fatal(err)
    }
	fmt.Printf("\nsecurity group created      %s", sgid)
	
	err = cloud.CreateInstProf()
	if err != nil {
		log.Fatal(err)
	}
	err = appendToFile("aws-resources.txt", "role                 ec2-admin-role-vaultAD")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nrole created                ec2-admin-role-vaultAD")
	
	err = appendToFile("aws-resources.txt", "\ninstance profile     ec2-InstProf-vaultAD")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("instance profile created    ec2-InstProf-vaultAD")
	
	fmt.Println("list of resources created   awsResources.txt")
	fmt.Println("output file created         ad-output.txt")
	// wait for instance profile to be created sometimes necessary to avoid not found error
	time.Sleep(5 * time.Second)

	pubDNS, err := cloud.BuildEC2()
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create("ad-output.txt")
	if err != nil {
		fmt.Printf("failed to create file: %v", err)
	}
	defer file.Close()

	output := fmt.Sprintf(`
Success! This output has been saved to ad-output.txt for reference.

Environment nearly ready. Server will need an additional few minutes to bootstrap AD even after connection established. 

LDAPS Enabled.
______________________________________________________________________

forest (root) dn    DC=vaultest,DC=com

root-user           Administrator
binddn              cn=Administrator,cn=Users,dc=vaultest,dc=com
password            admin

test-user           vaultusr01
dn                  CN=vaultusr01,CN=Users,DC=vaultest,DC=com
password            Hashi@pswd
______________________________________________________________________

ldapsearch example (not installed): 

ldapsearch -x -H ldap://%s:389 -D "cn=Administrator,cn=Users,dc=vaultest,dc=com" -w admin -b "dc=vaultest,dc=com" -s sub "(objectclass=user)"
______________________________________________________________________

Run this to connect: 

ssh -i key.pem -o StrictHostKeyChecking=no Administrator@%s
______________________________________________________________________

Then run the following to see AD details (if error, will need to wait a bit longer): 

> powershell
> Get-ADForest
> Get-ADUser -Filter *
> Get-ADUser -Identity Administrator -Properties *
______________________________________________________________________

Can also use ssh to run a single cmd: 

ssh -i key.pem -o StrictHostKeyChecking=no Administrator@%s 'powershell -Command Get-ADUser -Identity 'vaultusr01' -Properties *'
______________________________________________________________________

Additional powershell AD commands: 
# view existing user's full details:
Get-ADUser -Identity 'vaultusr01' -Properties *

# create a new user:
New-ADUser -Name 'vaultusr02' -AccountPassword (ConvertTo-SecureString -AsPlainText 'Hashi@pswd' -Force) -Enabled $true

# delete user without confirmation prompt:
Remove-ADUser -Identity 'vaultusr02' -Confirm:$false
`, pubDNS, pubDNS, pubDNS)

	_, err = file.WriteString(output)
	if err != nil {
		fmt.Printf("failed to write to file: %v", err)
	}

	cmd := exec.Command("cat", "ad-output.txt")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    err = cmd.Run()
    if err != nil {
        fmt.Println("Error printing output file:", err)
    }
}

func CleanupCloud() {
	if err := os.Remove("key.pem"); err != nil {
		fmt.Printf("key.pem file may not exist: %v\n", err)
	}
	if err := os.Remove("ad-output.txt"); err != nil {
		fmt.Printf("output file may not exist: %v\n", err)
	}
	if err := os.Remove("aws-resources.txt"); err != nil {
		fmt.Printf("output file may not exist: %v\n", err)
	}
	if err := os.Remove("ad-output.txt"); err != nil {
		fmt.Printf("output file may not exist: %v\n", err)
	}
	if err := cloud.TerminateEC2Instance(); err != nil {
		fmt.Printf("instance may not have been created: %v\n", err)
	}
	if err := cloud.DeleteKeyPair(); err != nil {
		fmt.Printf("key pair may not exist: %v\n", err)
	}

	if err := cloud.DetachPolicyFromRole(); err != nil {
		fmt.Printf("policy may not have been created: %v\n", err)
	}

	if err := cloud.DetachRoleFromInstanceProfile(); err != nil {
		fmt.Printf("error detaching role from instance profile: %v\n", err)
	}
	if err := cloud.DeleteInstanceProfile(); err != nil {
		fmt.Printf("error deleting instance profile: %v\n", err)
	}
	if err := cloud.DeleteRole(); err != nil {
		fmt.Printf("error deleting role: %v\n", err)
	}
	if err := cloud.DeleteSecurityGroup(); err != nil {
		fmt.Printf("error deleting security group: %v\n", err)
	}
}

func main() {
	if runCleanup {
		CleanupCloud()
	} else {
		bootStrap()
	}
}
