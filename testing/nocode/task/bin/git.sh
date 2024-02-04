#!/usr/bin/env bash

# Function to display help message
function display_help() {
    echo "Usage: $0 --config <config_file> [--name <project_name>] [--path <project_path>] [--command <command>] [additional_arguments]"
    echo "Options:"
    echo "  --config     Specify the YAML configuration file"
    echo "  --name       Name of the project (required)"
	echo "  --workspace  Path of the workspace (required)"
    echo "  --path       Path of the project (optional)"
    echo "  --command    Name of the command to execute (optional)"
    echo "  --help       Display this help message"
    exit 1
}

# Function to parse the YAML configuration file and extract project details
function parse_yaml() {
    local yaml_file="$1"
    local project_name="$2"

    if ! command -v yq &> /dev/null; then
        2>&1 echo "Please install 'yq' (a YAML parser) to use this feature."
        exit 1
    fi

    project_config=$(yq eval ".repos[] | select(.name == \"$project_name\")" "$yaml_file")

    if [ -z "$project_config" ]; then
        2>&1 echo "Project with name '$project_name' not found in the configuration file."
        exit 1
    fi
}

# Parse command line arguments
arguments=()
verbose=0
while [[ $# -gt 0 ]]; do
    case "$1" in
        -c|--config)
            config_file="$2"
            shift 2
            ;;
        -n|--name)
            project_name="$2"
            shift 2
            ;;
        -p|--path)
            project_path="$2"
            shift 2
            ;;
        -c|--command)
            command_name="$2"
            shift 2
            ;;
        -w|--workspace)
            workspace="$2"
            shift 2
            ;;
		-v|--verbose)
			verbose=1
			shift
			;;
        -h|--help)
            display_help
            ;;
		--) # End of all options, consume all remaining arguments as additional arguments for the command
			shift
			;;
        *)
            break
            ;;
    esac
done

# Check if a configuration file is provided
if [ -z "$config_file" ]; then
    2>&1 echo "You must specify a YAML configuration file using --config."
    display_help
fi

# Check if a project name is provided
if [ -z "$project_name" ]; then
	2>&1 echo "You must specify the name of the project using --name."
	display_help
fi

# Check if a workspace is provided
if [ -z "$workspace" ]; then
	2>&1 echo "You must specify the path of the workspace using --workspace."
	display_help
fi

# If project_name is provided, parse the configuration file to extract project details
parse_yaml "$config_file" "$project_name"

# Any subsequent failing commands will cause the script to exit
set -e

project_path=$(yq eval '.path' <<< "$project_config")

# Join the workspace and project path
project_path=$(realpath "$workspace/$project_path")

if [ -z "$command_name" ] && [ $# -gt 0 ]; then
	command_name="$1"
	shift
fi

if [ $verbose -eq 1 ]; then
	# Print the provided project details
	echo "Project Name: $project_name"
	echo "Project Path: $project_path"
	echo "Config File: $config_file"
	echo "Workspace: $workspace"
	echo "Command Name: $command_name"

	# Execute the specified command with additional arguments
	echo "Additional Arguments: $@"
fi

case "$command_name" in
	"status")
		git -C "$project_path" status $@
		;;
	"clone")
		project_url=$(yq eval '.url' <<< "$project_config")
		git clone "$project_url" "$project_path" $@
		;;
	"pull")
		git -C "$project_path" pull $@
		;;
	"fetch")
		git -C "$project_path" fetch $@
		;;
	"log")
		git -C "$project_path" log $@
		;;
	"push")
		git -C "$project_path" push $@
		;;
	"tag")
		git -C "$project_path" tag $@
		;;
	*)
		2>&1 echo "Invalid command: $command_name"
		exit 1
		;;
esac

# Exit the script