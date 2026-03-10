namespace ClientApp.Composition;

sealed class AppHost
{
    public AppHost()
    {
        ServiceCollection serviceCollection = new();
        Services = serviceCollection
            .AddClientApp()
            .BuildServiceProvider();
    }

    public IServiceProvider Services { get; }
}
