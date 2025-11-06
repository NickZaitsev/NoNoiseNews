# Rules for AI Bot

Follow these rules consistently across the entire codebase. Our primary goal is to implement our planned program effectively: ensuring it is scalable, secure, and free of errors, while following best practices and sound design patterns. If you need additional information, ask. If you identify a suboptimal implementation plan, propose a better alternative. If you encounter code that contains errors or poor patterns, stop and report it, and suggest a proper fix.

## 1. **Comments & Documentation**
- Only add comments where necessary — code should mostly be self-explanatory.  
- Use **docstrings** for all public functions, classes, and modules.  

## 2. **Functions & Methods**
- Functions should do **one thing well**; avoid large, multipurpose methods.  

## 3. **Dependencies**
- Use the minimal set of dependencies required.  
- Pin versions in lockfiles.  
- Avoid floating versions.  
- Remove unused dependencies.  

## 4. **Error Handling**
- Use explicit error handling.  
- Provide actionable and descriptive error messages.  

## 5. **Testing**
- All business-critical logic must be covered by tests.  
- Target at least **80% test coverage**.  
- No untested critical paths.  
- Tests must be deterministic and independent of external state.  

## 6. **Architecture & Reuse**
- If significant changes in architecture or features occur:  
  - Review **ALL** project files.  
  - Update memory files in `.kilocode\rules\memory-bank\` and `README.md`.  
- Before writing new code:  
  - Check whether functionality already exists in the codebase or in a well-maintained dependency.  
  - Never duplicate logic across modules — extract common functionality into a shared utility or library.  
  - Reuse must not compromise clarity: shared code should be understandable and well-tested.  

## 7. **Configuration**
- Do not hardcode values (API keys, database credentials, environment-specific settings).  
- Store configurable values in a central configuration file (e.g., `config.py`) or environment variables (`.env`).  
- Access configuration values through a single, well-defined interface.  

## 8. **Other Rules**
- Don’t write any new `.md` files.  
- Use efficient data structures.  

## 9. **Logging**
- Use structured logging instead of `print` statements.  
- Include context in log messages (function name, parameters, error info).  

## 10. **Security**
- Validate all external inputs to prevent injections.  
- Sanitize data before logging.  
- Follow secure authentication and authorization practices.  

## 11. **Dead Code & Unused Artifacts**
- Regularly review the codebase for unused functions, variables, classes, imports, and files.  
- Remove any code that is no longer used or referenced.  
- Ensure removed code does not break dependencies or shared functionality.  

## 12. **Refactoring & Code Quality**
- Refactor code proactively when you notice:  
  - Poor readability or confusing logic.  
  - Duplicated functionality.  
  - Inefficient algorithms or data structures.  
  - Violations of existing coding standards.  
- Ensure refactoring preserves functionality and passes all tests.  
- Prefer **incremental, small refactors** over large, risky changes.  

## 13. **Configuring**
- Whenever you add a new major feature or introduce configurable behavior, make it toggleable and centralized.
- Do not hardcode (“magic”) values directly in the code — put them in config.py and reference environment variables from .env.