import { DataState } from '../../shared/DataState';
import { StatusBadge } from '../../shared/StatusBadge';
import { useAsyncData } from '../../shared/useAsyncData';
import { useSession, useToken } from '../../shared/SessionContext';
import { readDashboard } from '../../services/dashboard_service';
import { listTechnician } from '../../services/technician_service';

export function DashboardPage() {
  const token = useToken();
  const { isAdministrator } = useSession();
  const { data, loading, error } = useAsyncData(() => readDashboard(token), [token]);
  const technicians = useAsyncData(
    () => (isAdministrator ? listTechnician(token) : Promise.resolve([])),
    [token, isAdministrator],
  );

  const busyTechnicians =
    data?.busyTechnician && data.busyTechnician.length > 0
      ? data.busyTechnician
      : (technicians.data ?? []).filter((item) => item.busy);

  return (
    <section>
      <h2 className="screen-title">Panel del taller</h2>
      <DataState loading={loading} error={error} empty={!data}>
        <div className="grid">
          <article className="stat-card">
            <div className="stat-card__value">{data?.openOrderCount ?? 0}</div>
            <div className="stat-card__label">Ordenes abiertas</div>
          </article>
          {(data?.statusCount ?? []).map((item) => (
            <article className="stat-card" key={item.status}>
              <div className="stat-card__value">{item.count}</div>
              <div className="stat-card__label">
                <StatusBadge status={item.status} />
              </div>
            </article>
          ))}
        </div>
        <section className="card">
          <h3 className="card__title">Tecnicos ocupados</h3>
          {busyTechnicians.length === 0 ? (
            <p className="state-message">No hay tecnicos ocupados.</p>
          ) : (
            <div className="table-scroll">
              <table className="data-table">
                <thead>
                  <tr>
                    <th scope="col">Tecnico</th>
                    <th scope="col">Orden</th>
                    <th scope="col">Placa</th>
                  </tr>
                </thead>
                <tbody>
                  {busyTechnicians.map((technician) => (
                    <tr key={technician.id}>
                      <td>{technician.fullName}</td>
                      <td>{technician.activeOrderNumber || 'En progreso'}</td>
                      <td>{technician.activeVehiclePlate || 'S/P'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </DataState>
    </section>
  );
}
