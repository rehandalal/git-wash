#!/bin/bash

print_2nl() {
  printf '%s\n\n' "$1"
}

git_prune_remote() {
  echo "Pruning..."
  git fetch --prune
  print_2nl "Complete! 🎉"
}

git_get_merged_local_branches() {
  git branch --merged | grep -E -v "(^\*|$1)"
}

git_remove_all_merged_local_branches() {
  echo "Removing all..."
  git_get_merged_local_branches "$1" | xargs -I % git branch -d %
  print_2nl "Complete! 🎉"
}

git_remove_each_merged_local_branch() {
  for branch in $(git_get_merged_local_branches "$1"); do
    read -rp "Do you want to remove branch: $branch [y/N]: " yn
    case $yn in
      [Yy]* ) git branch -d "$branch";;
      * ) print_2nl "Branch skipped...";;
    esac
  done
  print_2nl "Complete! 🎉"
}

git_remove_squash_merged_local_branches() {
  strategy=$1
  main_branch=$2
  current_branch=$3

  if ! git checkout -q "$main_branch";
  then
    echo "An error occurred! Unable to checkout main branch."
    print_2nl "Skipping..."
  else
    for branch in $(git for-each-ref refs/heads/ "--format=%(refname:short)"); do
      ancestor=$(git merge-base "$main_branch" "$branch")
      rp=$(git rev-parse "$branch^{tree}")
      ct=$(git commit-tree "$rp" -p "$ancestor" -m _)

      if [[ $(git cherry "$main_branch" "$ct") == "-"* ]]; then
        if [ "$strategy" == "EACH" ]; then
          read -rp "Do you want to remove branch: $branch [y/N]: " yn
          case $yn in
            [Yy]* ) git branch -D "$branch";;
            * ) print_2nl "Branch skipped...";;
          esac
        else
          git branch -D "$branch"
        fi
      fi
    done
    git checkout -q "$current_branch"
    print_2nl "Complete! 🎉"
  fi
}

git_remove_all_squash_merged_local_branches() {
  git_remove_squash_merged_local_branches "ALL" "$1" "$2"
}

git_remove_each_squash_merged_local_branch() {
  git_remove_squash_merged_local_branches "EACH" "$1" "$2"
}

main() {
  # Check if the working tree is dirty
  if [[ -n $(git status -s) ]]; then
    echo "Make sure your working tree is clean before attempting to run this script."
    exit 1
  fi

  # Check if a main branch has been configured for the given repo
  main_branch=$(git config --get git-wash.main-branch)

  if [ -z "$main_branch" ]; then
    echo "Uh oh, you have not configured a main branch for this repo!"
    while [ -z "$main_branch" ]; do
      read -rp "Name of main branch (e.g.: master): " main_branch
    done
    git config git-wash.main-branch "$main_branch"
    print_2nl "Main branch set to: $main_branch"
  fi

  # Get the current branch the user is on
  current_branch=$(git branch --show-current)

  # If the user is in detached HEAD state then use the current HEAD as the branch
  if [ -z "$current_branch" ]; then
    current_branch=$(git rev-parse HEAD)
  fi

  # Prompt the user to prune stale remote branches
  read -rp "Do you want to prune remote branches that are deleted or merged? [y/N]: " yn
  case $yn in
    [Yy]* ) git_prune_remote;;
    * ) print_2nl "Skipping...";;
  esac

  # Prompt the user to delete local branches that were merged
  read -rp "Do you want to remove local branches that are merged? [y/N]: " yn
  case $yn in
    [Yy]* )
      read -rp "Remove all merged branches at once? [y/N]: " yn;
      case $yn in
        [Yy]* ) git_remove_all_merged_local_branches "$main_branch";;
        * ) git_remove_each_merged_local_branch "$main_branch";;
      esac;
    ;;
    * ) print_2nl "Skipping...";;
  esac

  # Prompt the user to delete local branches that were squash-merged
  read -rp "Do you want to remove local branches that are squash-merged? [y/N]: " yn
  case $yn in
    [Yy]* )
      read -rp "Remove all squash-merged branches at once? [y/N]: " yn;
      case $yn in
        [Yy]* ) git_remove_all_squash_merged_local_branches "$main_branch" "$current_branch";;
        * ) git_remove_each_squash_merged_local_branch "$main_branch" "$current_branch";;
      esac;
    ;;
    * ) print_2nl "Skipping...";;
  esac
}

if [[ "${#BASH_SOURCE[@]}" -eq 1 ]]; then
    main "$@"
fi
