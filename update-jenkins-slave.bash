#!/bin/bash
set -eux


update_branch() {
    # Branch or pull a branch.
    local_branch=$1
    local_dir="$(basename $local_branch | cut -d ':' -f2)"
    local_path="$HOME/$local_dir"
    if [[ -d $local_path ]]; then
        bzr pull -d $local_path
    else
        bzr branch $local_branch $local_path
    fi
}


update_git_repo() {
    # Clone or pull a git repo.
    git_repo=$1
    local_dir="$(echo $git_repo|sed -r 's/.*\/([^\/]*)\.git/\1/')"
    local_path="$HOME/$local_dir"
    if [[ -d $local_path ]]; then
        (cd $local_path; git pull $git_repo)
    else
        git clone $git_repo $local_path
    fi
}


get_os() {
    # Get the to OS name: ubuntu, darwin, linux, unknown.
    local_uname=$(uname -a)
    if [[ "$local_uname" =~ ^.*Ubuntu.*$ ]]; then
        echo "ubuntu"
    elif [[ "$local_uname" =~ ^.*Darwin.*$ ]]; then
        echo "darwin"
    elif [[ "$local_uname" =~ ^.*Linux.*$ ]]; then
        # Probably CentOS.
        echo "linux"
    else
        echo "unknown"
    fi
}


# This works when the slave was setup by the jenkins-juju-ci subordinate
# charm, or when a person installed the keys in .ssh by links to
# cloud-city.
bzr --no-aliases launchpad-login juju-qa-bot

echo "Updating branches"
OS=$(get_os)
update_branch lp:workspace-runner
update_branch lp:juju-release-tools
update_branch lp:juju-ci-tools
update_branch lp:juju-ci-tools/repository
update_branch lp:~juju-qa/+junk/cloud-city
if [[ $OS == "ubuntu" ]]; then
    update_git_repo git@github.com:juju/hammer-time.git
fi

echo "Updating permissions"
sudo chown -R jenkins $HOME/cloud-city
chmod -R go-rwx $HOME/cloud-city
chmod 700 $HOME/cloud-city/gnupg
chmod 600 $HOME/cloud-city/staging-juju-rsa

echo "Updating dependencies from branches"
if [[ $OS == "ubuntu" ]]; then
    make -C $HOME/juju-ci-tools install-deps
    make -C $HOME/workspace-runner install
    if ! (lsb_release -c|grep trusty); then
        make -C $HOME/hammer-time develop
    fi
elif [[ $OS == "darwin" ]]; then
    $HOME/juju-ci-tools/pipdeps.py install
fi

echo "$HOSTNAME update complete"
