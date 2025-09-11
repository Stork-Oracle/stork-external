#!/bin/bash

STARTUP_ANIMATION_PATH="apps/lib/generate/frames"

function run_animation() {
    # Get base path (current working directory)
    base_path=$(pwd)
    frames_path="${base_path}/${STARTUP_ANIMATION_PATH}"

    # Check if frames directory exists
    if [ ! -d "$frames_path" ]; then
        echo "Error: Failed to read frames directory: $frames_path"
        exit 1
    fi

    # Pre-load all frames into an array to reduce I/O operations
    # Using sort -n to ensure numerical order
    declare -a frames
    i=0
    while IFS= read -r -d '' file; do
        frames[$i]="$(cat "$file")"
        ((i++))
    done < <(find "$frames_path" -type f -print0 | sort -z)

    # Display the pre-loaded frames
    for frame in "${frames[@]}"; do
        # Clear the screen
        echo -e "\033[H\033[2J"
        
        # Print the frame content from memory
        echo "$frame"
        
        # Use a more precise sleep command if available
        sleep 0.1
    done

    # Clear the screen at the end
    echo -e "\033[H\033[2J"
}

# Run the animation
run_animation
