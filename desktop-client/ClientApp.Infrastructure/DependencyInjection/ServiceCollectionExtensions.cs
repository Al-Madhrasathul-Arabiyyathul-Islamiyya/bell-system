using ClientApp.Application.Abstractions;
using ClientApp.Infrastructure.Audio;
using ClientApp.Infrastructure.Http;
using ClientApp.Infrastructure.Logging;
using ClientApp.Infrastructure.Persistence;
using ClientApp.Infrastructure.Realtime;
using ClientApp.Infrastructure.Services;
using Microsoft.Extensions.DependencyInjection;

namespace ClientApp.Infrastructure.DependencyInjection;

/// <summary>
/// Registers infrastructure-layer services for the desktop client.
/// </summary>
public static class ServiceCollectionExtensions
{
    /// <summary>
    /// Adds infrastructure implementations for application contracts.
    /// </summary>
    /// <param name="services">The service collection to update.</param>
    /// <returns>The updated service collection.</returns>
    public static IServiceCollection AddClientInfrastructure(this IServiceCollection services)
    {
        ArgumentNullException.ThrowIfNull(services);

        return services
            .AddSingleton<AuthenticatedHttpClientFactory>()
            .AddSingleton<AppLoggerAdapter>()
            .AddSingleton<RealtimeMessageRouter>()
            .AddSingleton<HttpScheduleApi>()
            .AddSingleton<WebSocketRealtimeUpdates>()
            .AddSingleton<SqliteLocalCache>()
            .AddSingleton<WindowsAudioPlaybackCoordinator>()
            .AddSingleton<AuthSession>()
            .AddSingleton<SystemClock>()
            .AddSingleton<IScheduleApi>(serviceProvider => serviceProvider.GetRequiredService<HttpScheduleApi>())
            .AddSingleton<IRealtimeUpdates>(serviceProvider => serviceProvider.GetRequiredService<WebSocketRealtimeUpdates>())
            .AddSingleton<ILocalCache>(serviceProvider => serviceProvider.GetRequiredService<SqliteLocalCache>())
            .AddSingleton<ILocalScheduleCache>(serviceProvider => serviceProvider.GetRequiredService<SqliteLocalCache>())
            .AddSingleton<IAudioPlaybackCoordinator>(serviceProvider => serviceProvider.GetRequiredService<WindowsAudioPlaybackCoordinator>())
            .AddSingleton<IAuthSession>(serviceProvider => serviceProvider.GetRequiredService<AuthSession>())
            .AddSingleton<IClock>(serviceProvider => serviceProvider.GetRequiredService<SystemClock>());
    }
}
