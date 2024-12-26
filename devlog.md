# Development Log

## 2024-03-XX - Package Reorganization

### Changes
- Split the monolithic package into separate packages:
  - `rpc`: Core RPC functionality
  - `session`: Session management
  - `stream`: Stream implementations (existing)
- Moved tests to their respective packages
- Updated all example code to use new package structure

### Benefits
- Better code organization
- More intuitive import paths
- Clearer separation of concerns
- Easier to maintain and test individual components

### Migration Notes
- Users need to update their imports to use the new package paths
- Session functionality is now accessed through the `session` package
- RPC functionality is now accessed through the `rpc` package
- No breaking changes in the API functionality

### Example Updates
- Updated example code to use new package structure
- Added better error handling using `Wait()`
- Improved session management examples 