# Project Context

This is a command line tool to provide a simple IPAM tool for allocating and managing CIDR blocks. It uses a file to store the superblock and allocated blocks, allowing for easy tracking, efficient use and management of IP address space. It also provides a markdown rendering of the allocated blocks, which can be useful for documentation or reporting purposes. The tool is designed to be simple and easy to use, while also providing powerful features for managing IP address space effectively.

## Key Concepts

- superblock: a CIDR block that represents the total available IP address space
- block: a CIDR block that is allocated from the superblock and can be assigned to a specific use or entity
- allocated blocks are tracked in a file, which allows for easy management and validation of the IP address space
- blocks are allocated in such a way to awoid as much as possible gaps within the IP space.

## Global Flags
- --file or -f : specify the file to use for storing the superblock and allocated blocks

## Commands

### touch
Check if a file exists and create it if missing.
#### Flags
- --name or -n : specify the name of the superblock
- --cidr : specify the CIDR of the superblock

### allocate
Allocate a block from the superblock and add it to the file.
#### Flags
- --size or -s : specify the size of the block to allocate (default is the block size defined in the superblock)
- --id or -i : specify an ID for the allocated block
- --output or -o : specify the output file (default is the input file)
- --description or -d : optional, specify a description for the allocated block

### release
Release a block and remove it from the file.
The block is identified by its CIDR or ID.

#### Flags
- --cidr or -c : specify the CIDR of the block to release
- --id or -i : specify the ID of the block to release

### cidr
Show the superblock CIDR or from a given block.
#### Flags
- --id or -i : specify the ID of the block to release
- --allocate or -a : if the block does not exist, allocate it
- --output or -o : specify the output file (used with --allocate, default is the input file)
- --size or -s : specify the size of the block to allocate (required with --allocate)
- --description or -d : optional, specify a description for the allocated block

### list
List all allocated blocks.

### validate
Validate the file and ensure there are no overlapping blocks.

### render
Render the allocated blocks in as a markdown table.
#### Flags
- --all-blocks or -a : include free blocks when rendering
- --render-format or -f : specify the format to render the table in (default is markdown)
