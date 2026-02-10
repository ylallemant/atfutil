# Project Context

This is a command line tool to provide a simple IPAM tool for allocating and managing CIDR blocks. It uses a file to store the superblock and allocated blocks, allowing for easy tracking and management of IP address space.

## Key Concepts

- superblock: a CIDR block that represents the total available IP address space
- block: a CIDR block that is allocated from the superblock and can be assigned to a specific use or entity
- allocated blocks are tracked in a file, which allows for easy management and validation of the IP address space
- blocks are allocated in such a way that the allocated IP space is as compact as possible, minimizing fragmentation and maximizing the efficient use of the available IP address space.

## Global Flags
- --file or -f : specify the file to use for storing the superblock and allocated blocks

## Commands

### allocate
Allocate a block from the superblock and add it to the file.
#### Flags
- --size or -s : specify the size of the block to allocate (default is the block size defined in the superblock)
- --id or -i : specify an ID for the allocated block
- --parent or -p : specify an ID for the parent block
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
- --parent or -p : specify an ID for the parent block (used with --allocate)
- --size or -s : specify the size of the block to allocate (required with --allocate)

### list
List all allocated blocks.

### validate
Validate the file and ensure there are no overlapping blocks.

### render
Render the allocated blocks in as a markdown table.
#### Flags
- --all-blocks or -a : include free blocks when rendering
- --render-format or -f : specify the format to render the table in (default is markdown)
