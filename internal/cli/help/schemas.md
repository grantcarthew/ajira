# ajira JSON Schemas

Field lists for --json output. `[]` denotes array response.

me: accountId, displayName, emailAddress, timeZone, active
project list: [id, key, name, lead, style]
board list: [id, name, type, project]
release list: [id, name, description, released, archived, releaseDate, startDate]
user search: [accountId, displayName, emailAddress, active]
field list: [id, name, custom, type]

issue list: [key, summary, status, statusCategory, type, priority, assignee]
issue view: key, summary, status, type, priority, assignee, reporter, created, updated, description, labels, project, attachments, comments
issue view attachments: [id, filename, size, mimeType, author, created, content]
issue view -c: comments contains [id, author, created, body]
issue create: key, id, self
issue edit: key, status
issue clone: originalKey, clonedKey, clonedId, linked, linkType
issue assign: key, assignee
issue move: key, status
issue move (list): [id, name, to.name]
issue delete: key, status
issue watch/unwatch: key, action
issue comment add: id, self, created
issue comment edit: id, self, created
issue attachment list: [id, filename, size, mimeType, author, created, content]
issue attachment add: issueKey, attachments[id, filename, size, mimeType, author, created, content]
issue attachment download: id, filename, size, output
issue attachment remove: issueKey, removed, count
issue type: [id, name, description, subtask]
issue status: [id, name, category]
issue priority: [id, name, description]

epic list: [key, summary, status, statusCategory, type, priority, assignee]
epic create: key, id, self
epic add: epicKey, issues, count
epic remove: issues, count

sprint list: [id, name, state, startDate, endDate, goal]
sprint add: sprintId, issues, count
