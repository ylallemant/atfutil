# Project Context

This is a command line tool to provide a simple IPAM tool.

## Key Concepts

- use file to store a superblock and a list of allocated blocks
- superblock contains the total CIDR and the block size
- blocks are allocated from the superblock and tracked in the file

## Commands

### allocate
Allocate a block from the superblock and add it to the file.
#### Flags
- --size or -s : specify the size of the block to allocate (default is the block size defined in the superblock)
- --id or -i : specify an ID for the allocated block
- --parent or -p : specify an ID for the parent block
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
- --parent or -p : specify an ID for the parent block (used with --allocate)
- --size or -s : specify the size of the block to allocate (required with --allocate)
- --in-place : modify the input file in place (used with --allocate)

### list
List all allocated blocks.

### validate
Validate the file and ensure there are no overlapping blocks.

### render
Render the allocated blocks in as a markdown table.
#### Flags
- --all-blocks or -a : include free blocks when rendering
- --render-format or -f : specify the format to render the table in (default is markdown)
