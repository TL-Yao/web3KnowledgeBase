# Project Guidelines

## Go Build Commands

This environment has GVM (Go Version Manager) configured in shell profile, which causes `cd` commands to fail with:
```
cd:1: command not found: __gvm_is_function
ERROR: GVM_ROOT not set. Please source $GVM_ROOT/scripts/gvm
```

**Solution**: Use absolute path to Go binary with `-C` flag instead of `cd`:

```bash
# WRONG - do not use cd
cd /path/to/backend && go build ./...

# CORRECT - use absolute Go path with -C flag
/usr/local/go/bin/go build -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend ./...
/usr/local/go/bin/go get -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend <package>
/usr/local/go/bin/go mod tidy -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend
```

The `-C` flag changes to the specified directory before executing the command, avoiding the shell `cd` issue.

## Project Structure

- `web3-insight/` - Main project directory
  - `backend/` - Go backend (Gin, GORM, Asynq)
  - `frontend/` - Next.js frontend with shadcn/ui
  - `docs/plans/` - Implementation plans
