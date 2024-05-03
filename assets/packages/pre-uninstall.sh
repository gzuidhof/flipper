#!/bin/sh

systemctl stop flipper || true
systemctl disable flipper || true
