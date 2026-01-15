# ajira JSON Schemas

Field lists for --json output.

## me

accountId, displayName, emailAddress, timeZone, active

## project list

Array: id, key, name, lead, style

## issue list

Array: key, summary, status, statusCategory, type, priority, assignee

## issue view

key, summary, status, type, priority, assignee, reporter, created, updated, description, labels, project, comments

comments (when -c used): id, author, created, body

## issue create

key, id, self

## issue comment add

id, self, created

## issue assign

key, assignee

## issue move

key, status

## issue move (list mode)

Array: id, name, to.name

## issue type

Array: id, name, description, subtask

## issue status

Array: id, name, category

## issue priority

Array: id, name, description

## issue watch/unwatch

key, action

## release list

Array: id, name, description, released, archived, releaseDate, startDate

## user search

Array: accountId, displayName, emailAddress, active

## field list

Array: id, name, custom, type
