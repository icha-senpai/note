import re

def find_milestone(repo, title, len=0):
    """Find the milestone in a repository that is similar to milestone title

    Args:
        repo (github.repository.Repository): The repository to search
        title (str): the title to match
        len: Version length limit; defaults to 0 for no limit.

    Returns:
        The milestone which title matches the given argument.
        If no milestone matches, it will return None
    """
    thisRelease = title.split("/")[-1]
    pat = re.search("v([0-9.]+)", thisRelease)
    if not pat:
        return None
    if len > 0:
        version = ".".join(pat.group(1).split(".")[:len])
    else:
        version = ".".join(pat.group(1).split(".")[:])

    for milestone in repo.get_milestones(state="all"):
        if version in milestone.title:
            return milestone

def generate_msg(desc_mapping, docmap):
    """Print changelogs from direction."""
    print()
    for header in docmap:
        if not desc_mapping[header]:
            continue
        print(f"#### {docmap[header]}\n")
        for item in desc_mapping[header]:
            print(f"* [{item['title']}]({item['url']})")
        print()

def get_issue_first_label(issue, docmap):
    """Get the first label from issue, if no labels, return empty string."""
    for label in issue.get_labels():
        if label.name in docmap:
            return label.name
    return ""

def generate_header_from_repo(repo_name, tag_name, lastestRelease, electron_version, action_file, HEADER=''):
    thisRelease = tag_name.split("/")[-1]
    pat = re.search("v([0-9.]+)", thisRelease)
    if not pat:
        return None

    return HEADER
