[33mcommit 9957b38985b680069dae9c393f4ecefaf1795f60[m[33m ([m[1;36mHEAD -> [m[1;32mfeature[m[33m, [m[1;31mupstream/main[m[33m, [m[1;31morigin/main[m[33m, [m[1;31morigin/HEAD[m[33m)[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Sun Mar 19 16:37:14 2023 +0530

    Update pull_request-template.md

[33mcommit b56a53e5ac24f70a4608728090f94ae90f318a3f[m[33m ([m[1;33mtag: v1.0.1-rc1[m[33m, [m[1;31mupstream/dev[m[33m)[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Sun Mar 19 00:15:37 2023 +0530

    [CLI][API] Added enhanced logging (#74)
    
    * Added the logging basics
    
    Info, Warn, Err
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Integrated the new print interface to the azure
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Modified the cli to use the new logger
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Credentials section to utilize the new logger
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Using New logger in SSHExecute function
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Added the verbose flag to all the subcommands
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Added notes and fixme for the firewall securities
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Fixed the unhandled err during adding more worker nodes
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Updated the installation script
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    ---------
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 9fdcab236e5e796f06dff0aaa9ce4fd5e24df9ba[m[33m ([m[1;33mtag: v0.2.1[m[33m)[m
Author: kranurag78 <84301342+kranurag78@users.noreply.github.com>
Date:   Tue Mar 7 23:13:37 2023 +0530

    [CI] update workflows for cosign 2.0 (#73)

[33mcommit 1837a4cbb777ba472c6ea78db4c4feb746cb09e8[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Mar 5 19:19:42 2023 +0530

    Removed the cosign signing
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 6c91781a1283ef2ec00e6814b169e516eb85e179[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Mar 3 21:41:11 2023 +0530

    Added the coverage report for azure api
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit af47e6f2746f415910b21c54652c29e9a5f7b3cf[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Fri Mar 3 21:42:53 2023 +0530

    [API](Azure) Adding support for Azure (#67)
    
    * Modularized the create and delete resource groups
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [CI] Fixed the windows build
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [CI] Updated the goversion in the build process
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [Docs] Updated the minimum version of Go for build process
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [API] Added Managed cluster feature with some TODOs
    
    - fetch the kubeconfig
    - have additional configurations
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [API](Azure) Added the Managed cluster support
    
    - left are some validation testing and adding test cases
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [BUG] Fixed the token userinput having echo enabled
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Removed the unused function
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [API][HA](Azure) support for ha clusters (#71)
    
    * Added the utility functions
    
    - Left SSH keypair and all
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [Bug] unable to unnmarshal state file
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Done with VM Create and Delete
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [API](Azure) Done with configs added loadbalancer
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [API](Azure) Done with infrastrcutre creation and deletion
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Done integration but issues with script exec
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * TODO fix the database not working
    
    issue with the username
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Added the worker node part as well
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [BUG] Fixed the database script
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Fixed the unable to join the controlplane nodes
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [HA](Azure) Done HA cluster delete, create, add, remove some workers
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    ---------
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [API](Azure) Done with the integration of API and CLI
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Added validation of region
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Added validation of nodesizes
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Added the validation for clusterName, region
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [API][CLI](Azure) Done with Azure support
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Added the testing for azure in gh actions
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Fixed the failing build script
    
    renamed package name
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Fixed the gh actions
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    ---------
    
    Signed-off-by: AvineshTripathi <avineshtripathi1@gmail.com>
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    Co-authored-by: AvineshTripathi <avineshtripathi1@gmail.com>

[33mcommit dd2a45b411b5580777b1a0cc7edc170da845ac54[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Feb 15 09:10:34 2023 +0530

    Removed the TODO for windows install script
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 33c4900d1a628acf0fb2f4579c5c1697e4f89e11[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Feb 15 09:07:47 2023 +0530

    [Docs] Added the windows script for oneline install
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 232f7e3c678e2fa1fdb3ebaf63ae797b01452063[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Feb 15 08:45:21 2023 +0530

    Added install script for windows
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit a65d0ac2d9765042eebb7134d1cb2c74fd3df802[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Tue Feb 14 22:05:04 2023 +0530

    [Docs] Added the single command install
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit d542d59ba11f319e45979480dc761f1f4fd80493[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Tue Feb 14 21:55:26 2023 +0530

    [Docs] Added the default install script for single line install
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 5e3fcf7a499e108b4b6ff1d9a3f88866088c51e6[m[33m ([m[1;33mtag: v0.1.1[m[33m)[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Tue Feb 14 14:41:56 2023 +0530

    Renaming of install scripts to builder scripts
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit db7f89730aa70699d78e14577cd6643789a4f26e[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Mon Feb 13 22:59:39 2023 +0530

    Update README.md

[33mcommit 4903e164ae734ce5ef8fba4226e72eb12c30a441[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Mon Feb 13 22:33:41 2023 +0530

    [CI](gh) added the release badge to readme
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 6884f117c93ae29e6ffd09b14a93041271f12fee[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Mon Feb 13 22:14:35 2023 +0530

    [CI](gh) added environment for gh action
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit fa3779312611743489f8ecb38ac1ca993112935f[m
Author: kranurag78 <84301342+kranurag78@users.noreply.github.com>
Date:   Mon Feb 13 21:45:53 2023 +0530

    [CI] [CLI] added gorelease process  (#69)
    
    add go release

[33mcommit 566090ba4d26c70f2e9e602f1dc26d50438511d7[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Mon Feb 13 20:17:42 2023 +0530

    [Docs] Added github action badges to readme

[33mcommit 48d43ab7282f46e9e7325b88d82fe32002856dd6[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Mon Feb 13 20:11:36 2023 +0530

    [Docs] Fixed the Readme
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 16aea2e97b144e829e4dc5e9d303c7f1f9cf3fe5[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Thu Feb 9 11:07:50 2023 +0530

    [CLI][API] Fixed the Kubeconfig printer and switch-context
    
    - Added windows based environment command to the KUBECONFIG export
    - fixed the shorthand for clustername in switch-context
      - from '-c' to '-n'
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 03f19305fe3df05b5b9b771519c57ff620a77997[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Thu Feb 9 09:45:30 2023 +0530

    Added different path output based on GOOS
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit a1ded5d39afd165812c80fced12b649ab4091758[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Wed Feb 8 21:50:40 2023 +0530

    [API][Test][Bug](Civo)(Local) Improved the Testcases and overall support (#66)
    
    * Added regex exp for clustername validation
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Added more testing for the utils for API
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [API,Test](CIVO) Improved the testcases and minor refactoring
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [API][Test](CIVO)(Local) Improved the testcases
    - civo
    - local
    
    [API] refactoring of ambiguous codebase in civo provider
    
    [BUG] Fixed the switch context sub-command
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [CI](gh) Updated the code coverage workflow
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    ---------
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 47d5b4716d57b103d3c3324d366d96f186ec3b16[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Tue Jan 31 23:56:05 2023 +0530

    Update README.md

[33mcommit 79c9e31edba732af6c8397204a1f30fb5d123c56[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Sun Jan 22 10:45:37 2023 +0530

    [Feature] [API](All) Support for SSH key-pair (#62)
    
    * [API](Civo) Addition of SSH key pair
    
    - added DeleteSSHKey
    - addded CreateSSHKey
    - added uploadSSH key
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Fixed the ssh-keygen command
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Fixed the failed SSH key deletion
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Added the ssh key pair auth
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Added method for hostkeyCallback as template
    
    - for future work
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Changed the SSH Exec from Password-based to SSH keypair-based
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Added restriction on clusterName to have only lowercase alphabets
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Fixed the testcases
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 027fe15f024ba3c021e8f771c33ad9ab8811ba5b[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Thu Jan 19 17:49:37 2023 +0530

    Updates to the maintainers readme
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit d18e4b834e303f96aa0095c1640e37b067f408a4[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Jan 15 20:20:06 2023 +0530

    Fixed the naming of cluster converted the region to lowercase
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 2cd56078a4edaad4f4ed09a3cc4df08420b7c71d[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Jan 15 20:03:36 2023 +0530

    [Feature][API](CIVO) Added note for CNI addition
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 40217668fe9e34f49c6d67c9b3fce84da50eb045[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Jan 8 22:32:14 2023 +0530

    Added the folder restriction for github actions
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 4cece2731c377770989a2b45fb77b1a5e2dc241f[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Sun Jan 8 22:27:52 2023 +0530

    Update issue templates

[33mcommit 5e59357506a3c51bfbb55258d5a486f7cef00fe5[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Sun Jan 8 22:23:03 2023 +0530

    Update issue templates

[33mcommit 6b2cc53bf520daa63c0bc8bc5e04f9904b56282e[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sat Jan 7 10:40:49 2023 +0530

    Added the new logo
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit c00c0e2d142dde1fe34700b1f9d19755d48edd78[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Tue Jan 3 23:46:28 2023 +0530

    [API,CLI](CIVO) Merging ha_civo/ and civo/ and credentials storage to json fmt (#59)
    
    * [API,CLI](CIVO) merging the civo/ and ha_civo/ folders
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Fixing the test scripts
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Did bug fixes and remapping of configurations
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * fixed Test cases failure
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Make credential storage as json objects
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Removed the comments
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit b8c49df788048571770ad761e6c6b36cf4e3cd98[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Mon Jan 2 11:15:32 2023 +0530

    [API] (testing) Addition test coverage report (#58)
    
    * Added coverage report
    * Added comment to pr
    * Added file change restrictions for coverage report
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 5bd693e8d339d9bcb92eec91a202008415de7c49[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Fri Dec 30 16:26:26 2022 +0530

    [API](CI) Added script to run testing for all providers (#57)
    
    * Added script to run testing for all providers
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Fixed the linux and macos unit testing
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 912d4381ddd7333c58e1ba153ee0540b7461887d[m
Author: Swastik Baranwal <swstkbaranwal@gmail.com>
Date:   Fri Dec 30 15:11:59 2022 +0530

    [API] (ALL) Optimised String appends (#56)
    
    * all: use strings.Builder instead of repeated += for strings
    
    * add tests and more optimization

[33mcommit dd4d243243bf31603f54c3ef4546f655a48171ee[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Tue Dec 27 23:13:04 2022 +0530

    Added the note
    do the insecure flag kubectl command for the first kubectl API call
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit daa7eb8ab6798a020a3c495cf4cbe9ea33a18a33[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Mon Dec 26 17:35:35 2022 +0530

    [CI,Testing](gh-actions) testing in different OS (#53)
    
    * [CI](Build) testing the build process in different environments
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Correction macos and windows commands
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * fixed the windows build process
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [CI](gh-actions) Improved the API testing
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [FIX](Testing) improved the getUsername
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Removal of Dockerfiles
    
    Issue #51
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Added triggers for specific path during APi testing
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Path triggers for testing builder
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 03c09fbdc9639718eef94302a224c6086745cc8f[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Mon Dec 26 14:08:53 2022 +0530

    [API](Refactoring) Rename of payload to utils
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit ae13930bb00db9156835879f66c3aaa271ded6cd[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Sun Dec 25 14:12:54 2022 +0530

    Update README.md

[33mcommit c66bab6114f83a51575c7622bbe58e84cc6ef49f[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Dec 25 14:05:57 2022 +0530

    [HA][CLI,API](CIVO) Added switch-context
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 86f5446e4f92ec878ccadec177109acf91f292d1[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Dec 25 13:53:56 2022 +0530

    [HA][CLI](CIVO) Integrated the HA API with CLI
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit a17c1adcefbbf95ff658fd1203f47d77517c4fc2[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Dec 25 11:24:53 2022 +0530

    Fixed the project structure
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 507196f469afa4130480c5e9e791ca07d45f6ca0[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Sun Dec 25 11:18:58 2022 +0530

    [HA](Civo) Addition of HA cluster API system (#37)
    
    * [Docs] Added Roadmap (#43)
    
    [Docs] Added Roadmap for the projects
    
    [HA](Civo) Added firewall creation
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    [API](Civo) Refactoring for Variable Naming
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    [HA,API](CIVO) Added SSH and Init scripts
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    UnResolved SSH issues
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    [HA](Civo) Modularize the structure
    
    know issues
    * insecure mode for ssh
    * sessionRun not working as expected
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Fixed the SSH issues
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    [HA](Civo) Removed the CIVO instance scripts dependency instead ssh into
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Marked the kubeadm as depricated
    
    it will be replaced by k3s
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Added error checks
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Added the single node k3s cluster script
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    first attempt of creating HA cluster
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Added the bash script for the ha setup
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Added a note ot update the kubeconfig with the public ip of loadbalcer
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Added Load balancer creation
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Integrated the Loadbalancer and Database
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Added controlplane and workerplane
    
    FIXME: workerplane not configuring
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Configured all some BUG found
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    All the control plane and worker nodes are working
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Fixed the BUG
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Refactroing and abstraction of implementation
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Using only one client for communicating to CIVO
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    FEARURE Done
    creation and deletion of HA cluster SUCCESS
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Added retry mechanism if ssh dial fails
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Added the recursive call to DeleteNetwork
    if some intance still available in the network
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Added name, region, nodeSize validation, retryCounter
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Moved the kubeconfig fetch before we join the worker nodes
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Replaced NGINX with HAPROXY for loadbalancing
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    More error handling cases
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Fixed the latency of response
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Fixed the scripts
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [FEATURE][API](CIVO) H.A. Cluster
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Fixed and Resolved the issues of HA cluster coding methods
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Added some emoji to related output messages
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [HA][API](CIVO) Added new configuration management via JSON
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [HA][API](CIVO) Removed the dependencies of worker node join for SSH
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [HA][API](CIVO) Added feature to add & delete more worker nodes
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Added warning message for delete cluster
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * Maked the instance to be public and private
    
    TO FIX: database instance is not becomming private instance
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    * [HA][API](CIVO) done with development Moved the feature to BETA
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit ca8a4fab96b9e2a3ac2016b091a59640ef80c4a8[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Nov 30 20:29:09 2022 +0530

    [Docs] moved the architecture diagrams
    
    from static to static/img in docs
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 27872594c871f29c09b8491a3314a2a16f4f78b6[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Nov 30 17:28:06 2022 +0530

    [Docs] removed HA for local setup
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 258eae34a7df36f93d72d7b6ee719b1d2127e397[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Fri Nov 25 11:28:27 2022 +0530

    [Docs] Added Roadmap (#43)
    
    [Docs] Added Roadmap for the projects

[33mcommit 970b55d964bf4d21799907acb28c0a4a77d2ba03[m
Merge: 69cfceb e9fb6ef
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Fri Nov 25 10:25:03 2022 +0530

    Merge pull request #40 from siddhant-khisty/readme-update
    
    [Docs]Addition of Usage steps in README

[33mcommit e9fb6ef68322c1496782d5d792e363a560f55e72[m
Merge: 24233a3 69cfceb
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Fri Nov 25 09:37:59 2022 +0530

    Merge branch 'main' into readme-update

[33mcommit 69cfcebc7b66a1a8bbd0e1ba57bb54037c0a44ac[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Nov 25 09:34:03 2022 +0530

    [CI](Docs) Restricted gh actions to docs folder
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 7d875601e1d20770dd43f28355db4390090b550b[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Nov 25 09:26:15 2022 +0530

    [Docs] Fixed the path problems
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 2d9f73d07ce5070a9ada6f7e0d20cb9caf0181c9[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Nov 25 08:53:58 2022 +0530

    [CI](github pages) Hugo site
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 26dc157a20c3a7ec7399bead5060bb98fed0d104[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Nov 25 00:03:19 2022 +0530

    [Docs] baseURL for hugo site
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 89649bd5536a916d68f85dbd43662fb90a2dc34e[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Thu Nov 24 23:56:48 2022 +0530

    Docs hugo server
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 24233a322fc361c40ad0541349cac488d614cf02[m
Merge: cf8ac42 9c11310
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Thu Nov 24 22:51:55 2022 +0530

    Merge branch 'main' into readme-update

[33mcommit cf8ac425ec68849b88bff6fe0250a0299d7b9ad4[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Thu Nov 24 18:36:40 2022 +0530

    Updated kubeconfig steps

[33mcommit 9c11310ea10647a757873ffb31f2ec4224a618b7[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Thu Nov 24 18:26:28 2022 +0530

    Delete README.md

[33mcommit 16c2d96eb5ca0720559388be66fb8b177cff1161[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Thu Nov 24 18:24:50 2022 +0530

    added index.html in docs

[33mcommit 59b5f4db90c0302dce91b50d922428f08b558f9a[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Thu Nov 24 18:19:28 2022 +0530

    added main markdown in docs

[33mcommit 25c6d027e661f49586cb8e73e1490afd7804e5c5[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Tue Nov 22 21:01:08 2022 +0530

    Update USAGE.md
    
    Co-authored-by: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>

[33mcommit d6fc19714c2c4b10aedf9201e9bbee6e68e79dc4[m
Merge: 0bdce42 e3f880f
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Mon Nov 21 19:19:36 2022 +0530

    Merge branch 'main' into readme-update

[33mcommit 0bdce424483595682288fd2d93815cee8944f9bc[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Mon Nov 21 18:58:54 2022 +0530

    Update USAGE.md
    
    Co-authored-by: Avinesh Tripathi  <73980067+AvineshTripathi@users.noreply.github.com>

[33mcommit 45bc033b835fba3cc0b71f4620f9b53aa445e4cd[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Mon Nov 21 18:58:31 2022 +0530

    Update USAGE.md
    
    Co-authored-by: Avinesh Tripathi  <73980067+AvineshTripathi@users.noreply.github.com>

[33mcommit 6ec2845d6abf7abf6c258ec8ac184da23edd0825[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Mon Nov 21 18:48:00 2022 +0530

    Update USAGE.md
    
    Co-authored-by: Avinesh Tripathi  <73980067+AvineshTripathi@users.noreply.github.com>

[33mcommit f0d9c5082d70e1c426f2c1a64cf36a5255099ff5[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Mon Nov 21 18:47:52 2022 +0530

    Update USAGE.md
    
    Co-authored-by: Avinesh Tripathi  <73980067+AvineshTripathi@users.noreply.github.com>

[33mcommit d6cc40da5dceee799742c1545b9353b3a637d778[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Mon Nov 21 18:46:49 2022 +0530

    Reframed usage

[33mcommit 7cde614fbfd92659210c8cc69ba787cf0bd6e80c[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Mon Nov 21 18:38:18 2022 +0530

    Highlight features

[33mcommit e281fb553c0cd91289701480882ea8257ffcdc89[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Mon Nov 21 18:36:27 2022 +0530

    Update README.md
    
    Co-authored-by: Avinesh Tripathi  <73980067+AvineshTripathi@users.noreply.github.com>

[33mcommit e3f880f5695806e7cbccc4806e3b17ea951f791d[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Nov 20 22:18:14 2022 +0530

    [CI](Docker) Marked Dockerfile(s) as DEPRICATED
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit f729a897b2eb6f46d127637d8e7e769cdab6fd95[m
Merge: da8defc 4ab8808
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Sat Nov 19 07:32:06 2022 +0530

    Merge pull request #41 from jonathino2590/jonathino2590-docs-Readme
    
    [Docs] Change of Contribution Guidelines

[33mcommit 4ab880829a1a6a37faa564a1256d8f3e7ff5d4f0[m
Author: Jonathan Lopez Torres <jonathino2590@gmail.com>
Date:   Fri Nov 18 14:09:41 2022 -0500

    [Docs] Change of Contribution Guidelines

[33mcommit b41214a1491bba1cb5c2f94756bd39805464f5ef[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Fri Nov 18 19:30:36 2022 +0530

    ADD: HA cluster as in progress

[33mcommit 08d463389573d5dbb0388acb6ddb8b9086775faa[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Fri Nov 18 17:39:01 2022 +0530

    Fix: Typo's in docs

[33mcommit ceca231bd206ae6a0b2ecf8e4eac8ba9c3173a31[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Fri Nov 18 17:35:26 2022 +0530

    Linked to install instructions

[33mcommit a8980dd5863fb1bf2a2459f5b87b23ce04af29a3[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Fri Nov 18 17:33:15 2022 +0530

    Remove linting error

[33mcommit 6df9ae01e0e4c0fdbd3277581c0fe5f09dcf51be[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Fri Nov 18 17:30:45 2022 +0530

    Change: Usage instructions in it's own file

[33mcommit 5f4be759120721cfc530ee48519d98d342953ba7[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Fri Nov 18 15:43:16 2022 +0530

    Update header

[33mcommit db80ad5e09dded26b85ace79530490a034ca08b4[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Fri Nov 18 15:41:51 2022 +0530

    ADD: Usage guide for ksctl

[33mcommit 76fab334fd4d70216cbe538780e698af9e21a782[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Fri Nov 18 14:38:31 2022 +0530

    Add: Instructions to register creds

[33mcommit 3487508f8e011c738cb4a7190cb0c13cfbbb4cf2[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Fri Nov 18 14:00:47 2022 +0530

    update contribution guide link

[33mcommit 07663523953ba47d073034354533ecb107d97a5d[m
Author: Siddhant Khisty <siddhantkhisty@gmail.com>
Date:   Fri Nov 18 13:48:22 2022 +0530

    update project scope

[33mcommit da8defc934c6a622156e44c568f7ad015af8b842[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Tue Nov 15 11:01:22 2022 +0530

    [CHANGELOG] fixed the verison numbering
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 9e92f8e40666b0a4ab31ff2cce1224fa02717441[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Nov 13 10:00:31 2022 +0530

    [HA](Civo) instance creation code added
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit a65a2194c8ab706d5a02323c22bbf24f5099104b[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Nov 13 00:18:33 2022 +0530

    [HA](Civo) Starting instance creation
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit f505d76ad64409d32018b4d19c31e59fa006c3e5[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Nov 11 22:34:31 2022 +0530

    [HA Cluster](Civo) template folder added
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit f4ba91a843241ad2317824e7e1d6206f90e2bc72[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Tue Nov 8 12:40:25 2022 +0530

    Update README.md

[33mcommit 7aecb818f168c25fa981d4d8b74d6cd8ea18ab4d[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Tue Nov 8 12:36:08 2022 +0530

    [Docs] Added cover images
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 17ce4e0cdd416683d02a7c3722b1a64b06b609d3[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Nov 6 21:58:41 2022 +0530

    [FIX] CLI
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 308b8c9b789b8f71b42f9c903c6e033cac7f4951[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Nov 6 21:33:33 2022 +0530

    [API](Civo)Fix the return value
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit be1126b692257a0a5d3970d6a2bca4dc38421ec6[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Nov 6 21:19:26 2022 +0530

    [CLI](Azure, AWS) Refactoring of init done
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 55eaf6eaf88df00dd2a9a9e75bfea67869cf1a76[m
Merge: 9aa47c2 1111e5e
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Sun Nov 6 21:12:34 2022 +0530

    Merge pull request #35 from kubesimplify/moveInitLogic
    
    [API](AWS,AZURE) Moved the Init credential to respective API handlers

[33mcommit 1111e5e6af0f405f0506dd6fcc094d59304c55f2[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Nov 6 21:09:38 2022 +0530

    [API](AWS,AZURE) Moved the Init credential to respective api handlers
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 9aa47c2c80283b06050367eab499273186ca86cf[m
Author: dipankar das <dipankardas0115@gmail.com>
Date:   Sat Nov 5 18:46:10 2022 +0530

    [Bug](Fix) Windows Support
    
    Fix #34
    
    Signed-off-by: dipankar das <dipankardas0115@gmail.com>

[33mcommit 2b097ff61172c0c975aa5669ca6251629265ec8b[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Thu Nov 3 22:08:33 2022 +0530

    Update README.md

[33mcommit 67b5ef6a7d33e84e75fd65914ff5de1b9016567d[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Thu Nov 3 22:07:22 2022 +0530

    [Script](Install)(Windows) Fixed the installation

[33mcommit a78babbf69ef16f84fce91474957afd6dc8b1b2b[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Thu Nov 3 21:45:15 2022 +0530

    [CLI] update the Reference API version
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit ca5e66cad633906087a3de77fdea7e83a6611e20[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Thu Nov 3 21:42:22 2022 +0530

    [BUG fix](API) Fixed the getEnv HOME
    
    it will try another option USERPROFILE if HOME is not available
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 5c8e748cc5a6cd0091c1216ee9873a4bdd9c7ff4[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Thu Nov 3 18:47:52 2022 +0530

    [API,CLI](Civo,Local) Added Support for Windows file Path
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit ba2333e784d45eb1bb3755fccad524e15289211a[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Thu Nov 3 18:37:01 2022 +0530

    [API](Civo,Local) Added the getPath as global scope
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit a2c55c43e29eafbbf72c3cf0b82fd901d1210da8[m
Merge: b869382 74a6089
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Thu Nov 3 18:29:32 2022 +0530

    Merge pull request #33 from kubesimplify/addSupportWin
    
    [API](Civo,Local) Added Windows File Path format

[33mcommit 74a60895c245d0edff0c559c20212658c63d5acd[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Thu Nov 3 18:22:15 2022 +0530

    [API](Civo,Local) Added Windows FilePath format
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit b869382534949e614dc0b91979f1703b48dc50db[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Tue Nov 1 08:45:10 2022 +0530

    [API,CLI](Civo,Local,Docker) Changed the config folder location
    
    From ~/.kube/ksctl/ to ~/.ksctl
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit b12579a38b56e11bad07b7a9c58723ee6d8fc6e3[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Tue Nov 1 08:38:34 2022 +0530

    [API] Changed the CLI config path
    
    Changed from ~/.kube/ksctl -> ~/.ksctl
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit a96abc85903f536c6efdcc370326e0d2e30bed9a[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Sat Oct 29 09:37:48 2022 +0530

    Create LICENSE

[33mcommit b216f573c5a42ccdeb6086cd0c8d7dbbbbee3207[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Thu Oct 27 11:47:11 2022 +0530

    Update README.md

[33mcommit 7053bfc725ebf25d9c56cfcd3c7f417bf45cf829[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Wed Oct 26 23:35:55 2022 +0530

    Update README.md

[33mcommit 7fa7877e9503a931efb8885f9b6ae12972c0c980[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Oct 26 23:12:04 2022 +0530

    [CLI](MacOS) added the support
    
    - [x] Issue #30
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit ecc60d5f92e9f3c946f7fe10e499e8daf1b8d0ca[m
Merge: b6db4b3 d1b7cd7
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Wed Oct 26 22:46:00 2022 +0530

    Merge pull request #32 from kubesimplify/supportMacOS
    
    [API,OS](MACOS) Added the API Support

[33mcommit d1b7cd7ec3939ea961a11764f4da533f5e734105[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Oct 26 22:37:52 2022 +0530

    [API,OS](MACOS) Added the API Support
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit b6db4b311d25a6295ec60d41388abf224ee79538[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Tue Oct 25 16:15:14 2022 +0530

    Update README.md

[33mcommit 4507a1ff0e3501ee86a417b118d2e4dbc780d28a[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Tue Oct 25 16:10:39 2022 +0530

    [Docs][CLI] Added support to Run the CLI inside docker container
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit d36b84523e699ebef120ca230a723b1ef9232fdc[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Tue Oct 25 14:16:30 2022 +0530

    [CLI,CI]Improved the Dockerfile for CLI
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 9cba1b76c3b5f1e26d640c0d9965fa7e7aec86ff[m
Merge: 1be91e8 808ab40
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Tue Oct 25 12:22:04 2022 +0530

    Merge pull request #31 from kubesimplify/dipankardas011-patch-1
    
    [CI] Create codeql.yml

[33mcommit 808ab40dd22999134bfe6a1eef2a1705eeda42c4[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Tue Oct 25 12:16:03 2022 +0530

    Create codeql.yml

[33mcommit 1be91e8563f6d901fbceab4acb4268cc664b827d[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Tue Oct 25 09:55:32 2022 +0530

    Added issue templates

[33mcommit c1b328662e68557e3dfb901dbf597f0c7e298dc5[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Tue Oct 25 09:45:11 2022 +0530

    Create CODE_OF_CONDUCT.md

[33mcommit 13c19b87a5d19a381d368cbafbe0b47ce5ea0168[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Tue Oct 25 09:32:38 2022 +0530

    Update CONTRIBUTING.md

[33mcommit e0c8d67d07f6357dabfdfb1e51277946797c1e4e[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Tue Oct 25 09:27:19 2022 +0530

    [Docs] Fixed the PR related queries
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit bd0d91f125036690df439363a2a8730499ec0d55[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Tue Oct 25 00:21:38 2022 +0530

    Rename CONTRIBUTION-GUIDE.md to CONTRIBUTING.md

[33mcommit 1b2cc1c09292cfecce712999ad3f793562d08884[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Tue Oct 25 00:14:14 2022 +0530

    [Docker] Added the CLI docker container
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit b7f887aa8cea5244fe826a31a09ea3e145f3b0d2[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Mon Oct 24 23:52:49 2022 +0530

    Renaming Done
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 804e3e2c1471927b2fe969186b99da0ffb9e1bec[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Mon Oct 24 10:50:15 2022 +0530

    Renaming Repo(I)
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit acfd5797f36a31dd8651832180885411cbf8508c[m
Merge: f9775a5 380b76d
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Mon Oct 24 10:11:52 2022 +0530

    Merge pull request #28 from kubesimplify/addGHActions
    
    [CI](GithubActions,Test) Added the github actions

[33mcommit 380b76d3d2808ef454c7404a1503ef64cf3cae77[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Mon Oct 24 10:08:16 2022 +0530

    [CI](GithubActions,Test) Added the github actions
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit f9775a5fd3f705573727f21afc1f3e219fb86554[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Mon Oct 24 09:45:07 2022 +0530

    Update CHANGELOG.md

[33mcommit 098c70cd36a62f7e661965425b9e5239c3222aea[m
Merge: 519a54c e2d491f
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Mon Oct 24 09:35:26 2022 +0530

    Merge pull request #26 from kubesimplify/AddSwitchCLI
    
    [CLI](Civo,Local) Added sub-command for switch-cluster

[33mcommit e2d491f1cbae52597caae99612e4ad55fdf2a88b[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Mon Oct 24 09:28:47 2022 +0530

    [CLI](Civo,Local) Improvement in scripting
    
    - switch cluster working
    - platform dependent executable
    - Readme update
    - makefile update
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 519a54cb892c0999407b98bcc45ed87dce811cf2[m
Merge: f4d0e9f 45a669d
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Mon Oct 24 09:03:03 2022 +0530

    Merge pull request #27 from kubesimplify/testCaseLocal
    
    [Test,API](Civo,Local) Added isPresent and validNodeSize

[33mcommit 45a669db6b1a2ee5caa692aaddf80f33a67eaf6a[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Mon Oct 24 08:58:19 2022 +0530

    [Test,API](Civo,Local) Added isPresent and validNodeSize
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit f4d0e9f9b56d3063dbca152a7d7bcea516a9720b[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Mon Oct 24 07:58:08 2022 +0530

    [BUG,API](PATCH)(Local) IsPresent Logical Error
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit fea0ee9217ff8c23d57e628307c06af8ec2ff25f[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Oct 23 23:21:52 2022 +0530

    [CLI](Civo,Local) Added sub-command for switch-cluster
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 0b66696f75d3213af1ea07c14a4a003c0f20ef1c[m
Merge: 02ef90f d94af47
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Sun Oct 23 22:29:45 2022 +0530

    Merge pull request #25 from kubesimplify/AddSwitch
    
    [API](Civo,Local) Added Switch-cluster Functionality

[33mcommit d94af47cdba72525cc72e77b480788cddecec45a[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Oct 23 22:21:10 2022 +0530

    [CLI](Civo,Local) Added Switch-cluster Functionality
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 02ef90f25a32cd1e16f183bd944e00df7b3b6e76[m
Merge: 697c1dc a0a211e
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Sun Oct 23 21:15:42 2022 +0530

    Merge pull request #24 from kubesimplify/AddGetCluster
    
    [Feature][CLI](Civo, Local) Added Get Clusters

[33mcommit a0a211e475c02afd650f3ec1e8d35a020500af05[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Oct 23 20:41:24 2022 +0530

    [CLI] Documentations for Srource
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 89dea853b8020b716e2f652914dbdfe6e7e71fe0[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Oct 23 09:40:27 2022 +0530

    [CLI] Added Get Cluster indent formating
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit c4e701f7826cfd13c8197a38cd3b795ebdbdd5c9[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Oct 23 09:32:03 2022 +0530

    [CLI](Civo, Local) Added Get Clusters
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 697c1dcef9607034e329620fa46dada96d364b55[m
Merge: 2b31449 b6628bf
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Sun Oct 23 08:47:04 2022 +0530

    Merge pull request #20 from kubesimplify/addLocalCLI
    
    [CLI](Local) Addition of CLI

[33mcommit b6628bf35beda5de8cc62e3194fc51fb6d49380d[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Oct 23 00:01:13 2022 +0530

    [CLI](Local) Removed the no of nodes as Required parameter
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 34845ca47ff2df38e060ebcaeeb60f0c3900bd96[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sat Oct 22 23:53:20 2022 +0530

    [PATCH,API](Civo,Local) Added
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 1d48a77e6a027c30c8c52ea94cb71b273ab1bde0[m
Merge: aecc18d 2b31449
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sat Oct 22 23:45:02 2022 +0530

    BUG fix merge #23

[33mcommit 2b31449ecef7b20120e50192b5ae0f9ccfbcf2c8[m
Merge: e4ecc82 47377c0
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Sat Oct 22 23:41:11 2022 +0530

    Merge pull request #23 from kubesimplify/bugFix01
    
    [BUG](CLI) Civo cluster's KUBECONFIG file save issue

[33mcommit 47377c0e8ff528ef9f054f96bf317466b1a5f3a5[m[33m ([m[1;31mupstream/bugFix01[m[33m)[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sat Oct 22 23:40:14 2022 +0530

    Added dev API main files for immediate testing purposes
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 32e5f6b732f347a1f2a08b14d7fda8cda9b22abc[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sat Oct 22 23:12:39 2022 +0530

    [BUG](CLI) Civo cluster's KUBECONFIG file save issue
    
    - PR #20
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit aecc18dd147060ff99127e1b3895fc99f8f292c9[m
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Fri Oct 21 20:38:12 2022 +0530

    [CLI](Local) Refactoring of createCluster flags
    
    Co-authored-by: Anurag  <81210977+kranurag7@users.noreply.github.com>

[33mcommit a447e345a256d80edf63649ed5623c228f250842[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Oct 21 19:36:37 2022 +0530

    [CLI](Local) Added the CLI configs
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit e4ecc82ac5d8e5ed334d7a9cb41b7464ed0da115[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Oct 21 19:15:53 2022 +0530

    [API](Local) Pretty Printing for KUBECONFIG ENV
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit ef51104c97b32cee8d4a7caf7042a1f7084b8190[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Oct 21 18:36:35 2022 +0530

    [CLI](Local) Updated the API Version
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit a40f84e7e27411b733b23393dd862c4ae84e6959[m
Merge: c12b298 7709f20
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Fri Oct 21 18:32:24 2022 +0530

    Merge pull request #12 from kubesimplify/addCreateLocal
    
    [API,Test](Civo, Local) Merged Local API

[33mcommit 7709f2019317ad9cc753e7ed21907162b12e2481[m[33m ([m[1;31mupstream/addCreateLocal[m[33m)[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Oct 21 18:29:12 2022 +0530

    [API,Test](Local,Civo) Ready for merging Local API
    
    - added testing scripts
    - fixed some unknowns
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 53b37be77ab69fc629ac030ca24974c27e0f105b[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Oct 19 14:44:46 2022 +0530

    [API](Local, Civo) Reorganizing & Refactoring structure
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit ae66ed6770e11f780e852d5954d93d49d4bfc312[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Oct 16 20:20:04 2022 +0530

    [CLI](Civo) Changed the hierachy of sub-commands
    now it has changed from kubesimpctl create-cluster -p [] ....
    to
    kubesimpctl create-cluster [] ...
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 3e4d2d7cdfb55ac15d7b571f13d55e0957c2c0a7[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Oct 16 20:17:22 2022 +0530

    [API,Testing](Civo) fixed the dependency issue
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 4a7f2a6c04c16a8ddb7a5069a11df88e633af416[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Oct 16 20:08:53 2022 +0530

    [API](Local, Civo) Create and delete completed
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit d8e3e9398d57682fefdc10ed493ceb853028e593[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Oct 16 13:47:50 2022 +0530

    [API](Local) MultiNode cluster
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 0f02300dd311156f84df57e964706ac1e941884f[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sat Oct 15 23:36:44 2022 +0530

    [API](Local) Refactoring of Create and Delete FuncUnit
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit b23dce7399e0e950295422f9560e61183d2fa22c[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sat Oct 15 19:42:32 2022 +0530

    [API](Local, Docker) Added demo for test Kind go packages
    
    - planning phase
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit c12b2981d69ef30bc83a6fefea5e188bc52b913b[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sat Oct 15 10:30:56 2022 +0530

    [CLI](Civo) Added API calls to CLI
    - improved the docs
    - install and uninstall
    - TODO: add option for additional applications and cni plugin
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 347b921c9ace090fc51452d7329b68e7087f7dc2[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sat Oct 15 10:28:09 2022 +0530

    [API](Bug) API_Fetch when no cred is present
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 09fbc6388947881743d680f3ad76c788cca14def[m
Merge: 3b10609 bd82acd
Author: Dipankar Das <65275144+dipankardas011@users.noreply.github.com>
Date:   Sat Oct 15 07:24:27 2022 +0530

    Merge pull request #3 from kubesimplify/addCreate_civo
    
    [API](Civo) Addition of create and deletion of cluster

[33mcommit bd82acd37eb8c20db0a01ab0fddea38fb58825e0[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Oct 14 19:54:45 2022 +0530

    [Docs](Civo)(Create)(Delete) Added function docs
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit d5ac369fb15aa7bd243560ee6d457354a380b9c5[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Oct 14 19:12:42 2022 +0530

    [API](Civo) Improvement to API
    
    **Added**
    - node size validator
    - applications, CNI plugin
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 01c9993704de18884af24078be4c9d5268615c12[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Oct 14 14:33:46 2022 +0530

    [Code](Kubeconfig) Improved the create cluster
    - Added preCluster check whether its already created
    - Kubeconfig printing options
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit debd6ff08243ee2a47c3653019cc046eac8c5eb7[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Thu Oct 13 23:28:21 2022 +0530

    [Test](Unit testing Civo) Added basic test cases
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit e9639f3819eba35bc0f68dd118996f509d5cf8a0[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Thu Oct 13 17:18:04 2022 +0530

    [Docs](API) Refactoring function arguments
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 000745cbc46169ea0d06666a87bbd59fbc4a8fef[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Oct 12 22:22:32 2022 +0530

    [CIVO](create&delete) Added the create and delete
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 746fd7b12d242ac9aeca7d23a8b1110a724fd4f7[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Oct 12 16:27:25 2022 +0530

    [CIVO](create&delete) Iteration I
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 3b1060939ba4cb0ae7f8f0c37ea3107d89740fa3[m
Author: Anurag <contact.anurag7@gmail.com>
Date:   Wed Oct 12 04:14:03 2022 +0530

    add global aliases for CLI
    
    Signed-off-by: Anurag <contact.anurag7@gmail.com>

[33mcommit 227c678a23ecd91d83d8372f2da8567447209f9b[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Mon Oct 10 21:31:58 2022 +0530

    [CLI](create-cluster) Addition of more parameters
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit cb10d551e9e44bc358b28d1cae2d7de639b94588[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Mon Oct 10 21:16:06 2022 +0530

    [API](Spec) defination improvement
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 9d422ba5c6a9145690b1a7da57430bf5a41733d0[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Mon Oct 10 08:39:09 2022 +0530

    [CLI, Docs](Improvement, Modules) Fixed the module hash used
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit b1b33c1066ca20d46115b23a9fc6becc9d4c6dcc[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Oct 9 11:36:46 2022 +0530

    Resolving the import naming
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 98f52c2ba952c819041bfb1ab6c21d1a2827d8a0[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Oct 9 11:25:50 2022 +0530

    Refactoring the package name
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit b7c1e411a03f28bb4975e476e05af8049e65b9bc[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Oct 7 23:30:09 2022 +0530

    Minor tweaks to ctl init
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 05d557af6c032133f3ef59d8d918043627af4387[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Oct 7 17:02:03 2022 +0530

    [Docs] Refractoring proposal
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 2b9c7123c58cdb31d6314727c528069045d3d063[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Thu Oct 6 18:05:24 2022 +0530

    Basics structure for API and cli done
    - init commands is almost done
      TODO: cloud verification
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit fbe365101107865d87f22277044a0db6414aa384[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Thu Oct 6 17:36:43 2022 +0530

    Refactoring API
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit a6c9e300f13d5f48394b49a30df86d3009a8598c[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Oct 5 22:36:21 2022 +0530

    Completed the first iteration of prototype
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 59f443eb4a0742b0512c9cb72649627854bc153e[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Oct 5 22:19:40 2022 +0530

    v0.0.2 API COMPLETED
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 4dab25961278dc868180520ca4f7cd721c97a797[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Oct 5 21:08:21 2022 +0530

    v0.0.1 API COMPLETED
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 40dd905f0f7050a3f591096b36f5f27da769b918[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Oct 5 10:25:31 2022 +0530

    Done success package call
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit be5687ac556b69e0b894ec3159e55d651ca3cf6b[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Oct 5 08:46:48 2022 +0530

    Added scripts
    - Install & Uninstall the CLI
    - Updated readme and makefile
    - Refactoring the package structure
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit bc42c085f7d1f6f774a9049f31c49b93433098d5[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Tue Oct 4 23:47:52 2022 +0530

    Improved the documentations
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 4f22b7fbb6b1a5d80dfb092c767b11010e386843[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Tue Oct 4 12:18:25 2022 +0530

    Added the folder structure for API Handlers
    
    api/eks, api/aks, api/local
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 49ea682d424cf4607dabc5a6db2c73e7734caa2e[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sun Oct 2 17:01:15 2022 +0530

    Refined the documentations
    
    * Changes to proposals
    * Makefile
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 1ce8ad5bfbaf08449ea28a9c8ef8320b2796b3bd[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Sat Oct 1 23:16:17 2022 +0530

    Added descriptions for the basic sub-commands
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit a9494afb74d5d52b87b32627a49ed047a2031201[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Fri Sep 30 15:23:50 2022 +0530

    Added Start&Stop clusters CLI
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 1c4cfdde6b05b47d7e3ed62488a6ff134562c0e9[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Sep 28 22:36:00 2022 +0530

    Added Initial Proposal
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>

[33mcommit 6a7b70ab1dd4c4963df164decc930ff5f7a5cd6c[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Sep 28 22:09:55 2022 +0530

    Improved the cli

[33mcommit 2ed214fcbf6a275cd75b14f13e0476fea11e802c[m
Author: Dipankar Das <dipankardas0115@gmail.com>
Date:   Wed Sep 28 22:02:56 2022 +0530

    Init done
    
    Signed-off-by: Dipankar Das <dipankardas0115@gmail.com>
