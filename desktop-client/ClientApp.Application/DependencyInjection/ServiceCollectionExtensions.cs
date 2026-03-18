using ClientApp.Application.State;
using ClientApp.Application.UseCases;
using ClientApp.Application.ViewModels;
using Microsoft.Extensions.DependencyInjection;

namespace ClientApp.Application.DependencyInjection;

/// <summary>
/// Registers application-layer services for the desktop client.
/// </summary>
public static class ServiceCollectionExtensions
{
    /// <summary>
    /// Adds application state, view models, and use cases.
    /// </summary>
    /// <param name="services">The service collection to update.</param>
    /// <returns>The updated service collection.</returns>
    public static IServiceCollection AddClientApplication(this IServiceCollection services)
    {
        ArgumentNullException.ThrowIfNull(services);

        return services
            .AddSingleton<ScheduleState>()
            .AddSingleton<ConnectionState>()
            .AddSingleton<PlaybackState>()
            .AddTransient<MainViewModel>()
            .AddTransient<MainShellViewModel>()
            .AddTransient<DashboardViewModel>()
            .AddTransient<ScheduleListViewModel>()
            .AddTransient<PlaybackStatusViewModel>()
            .AddTransient<LoginViewModel>()
            .AddTransient<SettingsViewModel>()
            .AddTransient<RefreshScheduleUseCase>()
            .AddTransient<StartRealtimeSyncUseCase>()
            .AddTransient<TriggerManualPlaybackUseCase>()
            .AddTransient<PauseBellSystemUseCase>();
    }
}
