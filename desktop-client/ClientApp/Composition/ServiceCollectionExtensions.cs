using ClientApp.Application.DependencyInjection;
using ClientApp.Infrastructure.DependencyInjection;

namespace ClientApp.Composition;

static class ServiceCollectionExtensions
{
    public static IServiceCollection AddClientApp(this IServiceCollection services)
    {
        ArgumentNullException.ThrowIfNull(services);

        return services
            .AddClientApplication()
            .AddClientInfrastructure()
            .AddTransient<ShellPage>()
            .AddTransient<MainPage>()
            .AddTransient<SettingsPage>();
    }
}
