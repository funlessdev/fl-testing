# Funless Testing Repo

This repo contains utilities and projects to setup a testing environment for the funless platform.

In the `vagrant` folder you can use vagrant to spin up some VMs to deploy and test funless,
running: `vagrant up --provider=libvirt`.

It spins up 3 VMs with docker installed on them and fl-cli installed.

To remove them use `vagrant destroy -f`. (the -f is to skip confirming.)