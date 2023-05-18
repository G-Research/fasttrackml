const openInNewTab = (url: string) => {
    window.open(url, '_blank', 'noreferrer');
};

export const openNewIssue = (base_url: string, title: string = "", body: string = "", assignee: string = "", labels: string[] = []) => {
    const url = new URL(base_url);
    url.searchParams.append("title", title);
    url.searchParams.append("body", body);
    url.searchParams.append("assignee", assignee);
    labels.forEach(label => url.searchParams.append("labels", label));
    openInNewTab(url.toString());
}

export const openNewEmail = (email: string, subject: string = "", body: string = "") => {
    const url = new URL(`mailto:${email}`);
    url.searchParams.append("subject", subject);
    url.searchParams.append("body", body);
    openInNewTab(url.toString());
}
