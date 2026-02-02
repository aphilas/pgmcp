#!/bin/bash
set -e

pg_restore -U postgres -d dvdrental /seed/dvdrental.tar
