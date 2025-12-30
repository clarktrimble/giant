
# Big picture thoughts:

  ## What giant does well:
  - Clean middleware pattern (Use() + tripper interface) - composable, testable
  - Config-driven - 12-factor friendly, struct tags for docs
  - Layered convenience: Send → SendJson → SendObject - pick your abstraction level
  - Doesn't over-abstract stdlib - http.Client and http.Response stay visible
  - Good package separation - logrt/statusrt/basicrt are independent, reusable

  ## Friction points I see:
  1. Config.Headers as []string - awkward pairs, could be map[string]string
  2. No status code access from SendJson/SendObject - you get data or error, but what if you want the status on success?
  3. Query params - no ergonomic way to add them; manual path building
  4. noRedirect is hardcoded - opinionated, some users want redirects
  5. SendObject error opacity - can't tell HTTP failure from unmarshal failure

  ## Missing patterns that many clients want:
  - Retries with backoff
  - Circuit breaker
  - Request/response hooks beyond logging

  Existential question: Is giant a general HTTP client or specifically for JSON APIs? The name doesn't say. If it's JSON-focused, lean into that. If general, the Json/Object methods feel bolted on.


  ## It's a JSON API client. To lean in:

  1. Typed responses with status
  ```
  type Response[T any] struct {
      Status  int
      Headers http.Header
      Body    T
  }

  func (g *Giant) Get[T any](ctx context.Context, path string) (Response[T], error)
  ```

  2. Distinguish error types
  ```
  type APIError struct {
      Status int
      Body   []byte  // raw response for inspection
  }

  func (e *APIError) Error() string { ... }

  // Now callers can:
  if apiErr, ok := err.(*APIError); ok {
      if apiErr.Status == 404 { ... }
  }
  ```

  3. REST verbs as methods
  ```
  giant.Get(ctx, "/users/123", &user)
  giant.Post(ctx, "/users", newUser, &created)
  giant.Delete(ctx, "/users/123")
  ```

  4. Query params
  ```
  giant.Get(ctx, "/search", &results,
      giant.Query("q", "term"),
      giant.Query("limit", "10"),
  )
  // or
  giant.Get(ctx, "/search?q=term&limit=10", &results)
  ```

  5. Path building
  ```
  giant.Get(ctx, giant.Path("/users/%s/posts", userID), &posts)
  ```

  The generics one (Response[T]) is the big win - no more passing &rcvObj and wondering if it worked.
