# mktodo

This project results in a golang based tool used to maintain TODO in
an existing markdown file.

Orientation the tool always works from within a git repo. If not
exists with a warning.

# Behavior

* Only the todo items are considered to be maintained. This means
  all the other content is left unchanged. This means that any
  content in other parts of the document are left exactly in place.
* If a document is mentioned but not created the user is prompted
  unless a command line switch has been given.
* A todo item is also called a topic
* A todo item is considered to be child of a project
* projects can be nested (e.g. "Technic" is a child of "Lego project
  targets" (see below)
* if no project is specified the parent is considered ``nil`` and
  the title is considered ``TODO``
* If the titles are not mentioned yet in the target document the
  default action is to append the needed content only when adding
  new topics
* if a topic is prefixed with "fixme:" or "FIXME:" then it is
  treated as a special case, highlighted in output (when not
  completed) and prioritized in listing.

# Coding

* Language is golang
* module name is ``github.com/jvzantvoort/mktodo``
* Preferred libraries
  * github.com/spf13/viper - configuration file handling
  * github.com/spf13/cobra - command line arguments
  * github.com/charmbracelet/bubbletea - tui handling
* adhere to ``go vet`` and ``golangci-lint`` checks
* aim for a high level of test coverage and quality
* defer text blocks to ``github.com/jvzantvoort/mktodo/messages``
  using a ``go:embed`` construct. 
* Avoid unnecessary code complexity
* Fill out the github workflows to ensure a complete ci/cd picture

# Configuration

In the root of the git repo a target called ".mktodo.yml" maintains
the information to work.

## Simple todo list

The project name ``default`` indicates the root of the document and
implies the title has a level 1 header (``#``).

``.mktodo.yml`` content:

```yaml
---
todofile: README.md
projects:
  - name: default
    title: TODO
    parent: nil
```

``README.md`` content:

```
# TODO

- [X] topic 1
- [ ] topic 2
```

## Simple non-default todo list

On the command line the project is refered to as ``lego`` in the
document it is displayed as ``# Lego project targets``. This because
in the configuration the ``parent`` is ``nil`` and the title is
``Lego project targets``.

``.mktodo.yml`` content:

```yaml
---
todofile: README.md
projects:
  - name: lego
    title: Lego project targets
    parent: nil
```

``README.md`` content:

```
# Lego project targets

- [X] topic 1
- [ ] topic 2
```

## Nested todo list

``.mktodo.yml`` content:

```yaml
---
todofile: README.md
projects:
  - name: lego
    title: Lego project targets
    parent: nil
  - name: technic
    title: Technic
    parent: lego
```

``README.md`` content:

```
# Lego project targets

## Technic

- [X] topic 1
- [ ] topic 2
```

# Interfacing

## Comand line interfacing

The command accepts the following subcommands:

* ``add`` or ``create`` add a new todo item
* ``remove``, ``rm`` or ``destroy`` remove a todo item
* ``done`` or ``complete`` set a todo item to done
* ``list`` or ``ls`` list the todo items
* ``open`` ``ls -o`` list the open todo items
* ``report`` create a graphical report of the todo items per
  project.
* ``tui`` interactive session that allows for the listing, updating,
  completion of todo items

### Add

Example command line uses:

Add ``topic 3`` to the default ``# TODO`` section:

```
mktodo add topic 3
```

Add ``topic 4`` to the ``# Lego project targets`` section:

```
mktodo add -p lego topic 4
```

Add ``topic 5`` to the ``## Technic`` section:

```
mktodo add -p lego.technic topic 5
```

### Remove

Remove ``topic 3`` from the default ``# TODO`` section:

```
mktodo rm topic 3
```

Remove ``topic 4`` from the ``# Lego project targets`` section:

```
mktodo rm -p lego topic 4
```

Remove ``topic 5`` from the ``## Technic`` section:

```
mktodo rm -p lego.technic topic 5
```


## Interactive menu (charmbracelet)

When called a listing is provided of projects and topics

Selecting a topic and pressing "e" will allow the user to edit the
topic content.

Selecting a topic and pressing space will toggle the status of the
topic from "todo" to "done" or back.

Selecting a topic and pressing "d" will present a pop-up asking if
the user is sure. Upon pressing "y" the topic is deleted.

Selecting a topic and pressing "-" will move the topic up the list.

Selecting a topic and pressing "+" will move the topic down the list.

Selecting a topic and pressing "x" will remove the topic from the
list but placing the contents in memory.

Selecting a project and pressing "p" will append the topic to the
list of topics in the project.

Selecting a project and pressing "a" will present the user with a
text field allowing for the creation of a new topic.

Pressing ``Esc`` will create a popup with the following options:
* ``q``, quit. When there are open changes the user is asked for confirmation and when given the
  application is exited without saving changes and if no changes are
  open the application is simply quit.
* ``s``, save. The current changes are save to the file(s)
* ``x``, save + quit. The current changes are save to the file(s)
  and the application is quit







