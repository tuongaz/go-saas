## USAGE


```go
app.OnBeforeServe().Add(func(ctx context.Context, e *app.OnBeforeServeEvent) error {
    sch, err := scheduler.New()
    if err != nil {
        return err
    }
    sch.Start()

    if _, err := sch.RunCron("*/1 * * * *", func() {
        log.Println("every second")
    }); err != nil {
        return err
    }

    return nil
})
```