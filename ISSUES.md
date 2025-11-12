# Issues

In this file the entries (issues) record newly discovered requests or changes, with their outcomes. No instructive content lives here. Read @NOTES.md for the process to follow when fixing issues.

Read @AGENTS.md, @ARCHITECTURE.md, @POLICY.md, @NOTES.md, @README.md and @ISSUES.md. Start working on open issues. Work autonomously and stack up PRs.

## Features  (100–199)

- [ ] [PN-100] Add a front-end.
  Notes:
    - The front end is a stand alone web app. It uses Google Sign and TAuss backend for JWT generation and authentication
    - The front end has two pages: 
    1. The landing page, which carries marketing funbction and also has a login button
    2. The main page only accessible after the login:
      - The main page displays a table of all the received messages. it has a column for status: delivered/queued, a time of dlivery, a sender
      - The table allows editing of queued messages: changing the time of delivery or cancelling the delivery. The status column will have a cancelled and errored statuses.
      - The main page has a header and the footer from the mpr-ui library (@tools/mpr-ui)
    - Styling: 
      1. Use footers and headers and stylign from mpr-ui (@tools/mpr-ui)
      2. Support theme switch in the footer
      3. Have a landing page
    - Backend
    1. Integrate with TAuth service for the front end JWT verification (use the same secret). The TAuth service and its documentation is under @tools/TAuth. The landing page uses TAuth client to wrap the sign in requests
    2. The main page uses TAuth client to keep the user signed in
    3. Prepare docker compose orchestration
  Deliverable: perform the code analysis and prepare an implementation plan. AS we are integrating with two packages: mpr-ui and Tauth, be sure to consider the plan of integration. Deliver a plan expressed as small atomic tasks in @ISSUES.md

## Improvements (200–299)

## BugFixes (300–399)

## Maintenance (400–499)

## Planning
*do not work on these, not ready*

