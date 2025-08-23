#!/bin/bash

LOG_DIR="$HOME/.termlogger/logs"
TEST_LOG_DIR="./testing/logs/"


if [ -d "$LOG_DIR" ]; then  
    read -p "Do you want to delete all of the main logs? (y/n) " -n 1 -r
    echo "" 
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🗑️ Deleting contents..."
        find "$LOG_DIR" -mindepth 1 -delete
        echo "✅ Deletion complete."
    else
        echo "⏩ Skipping deletion."
    fi
else
    echo "❌ Directory not found. Nothing to do."
fi

if [ -d "$TEST_LOG_DIR" ]; then
    read -p "Do you want to delete all of the testing logs? (y/n) " -n 1 -r
    echo "" 
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🗑️ Deleting contents..."
        find "$TEST_LOG_DIR" -mindepth 1 -delete
        echo "✅ Deletion complete."
    else
        echo "⏩ Skipping deletion."
    fi
else
    echo "❌ Directory not found. Nothing to do."
fi
