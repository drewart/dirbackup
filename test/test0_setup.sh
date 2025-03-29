#!/bin/bash

mkdir -p src/a
mkdir -p src/b
mkdir -p src/c
mkdir -p trg/a

echo "Hello, World a_old!" > trg/a/a.txt
echo "Hello, World delete!" > trg/a/z.txt
sleep 1
echo "Hello, World a!" > src/a/a.txt
echo "Hello, World b!" > src/b/b.txt
echo "Hello, World c!" > src/c/c.txz
