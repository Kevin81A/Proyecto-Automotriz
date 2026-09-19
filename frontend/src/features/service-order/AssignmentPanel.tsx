/* Allocation panel: pick an available technician for this order. A busy
   technician is shown with its badge and the backend rejection is surfaced as
   the Spanish message it returns. */
import { useState } from 'react';

import { ApiError } from '../../services/api_client';
import { assignTechnician, findAssignment } from '../../services/service_order_service';
import { listTechnician } from '../../services/technician_service';
import { DataState, ErrorBanner, SuccessBanner } from '../../shared/DataState';
import { AvailabilityBadge } from '../../shared/StatusBadge';
import { useAsyncData } from '../../shared/useAsyncData';
import { useSession, useToken } from '../../shared/SessionContext';

interface AssignmentPanelProps {
  serviceOrderId: string;
  isDelivered?: boolean;
  onChange: () => void;
}

export function AssignmentPanel({ serviceOrderId, isDelivered = false, onChange }: AssignmentPanelProps) {
  const token = useToken();
  const { isAdministrator } = useSession();
  const technician = useAsyncData(
    () => (isAdministrator ? listTechnician(token) : Promise.resolve([])),
    [token, isAdministrator],
  );
  const assignment = useAsyncData(
    () =>
      findAssignment(token, serviceOrderId).catch((failure: unknown) => {
        if (failure instanceof ApiError && failure.status === 404) {
          return null;
        }
        throw failure;
      }),
    [token, serviceOrderId],
  );
  const [technicianId, setTechnicianId] = useState('');
  const [error, setError] = useState('');
  const [confirmation, setConfirmation] = useState('');
  const [sending, setSending] = useState(false);

  const assignedTech = (technician.data ?? []).find(
    (t) => t.id === assignment.data?.technicianId,
  );

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError('');
    setConfirmation('');
    setSending(true);
    try {
      await assignTechnician(token, serviceOrderId, technicianId);
      setConfirmation('Tecnico asignado.');
      technician.reload();
      assignment.reload();
      onChange();
    } catch (failure) {
      setError(failure instanceof ApiError ? failure.message : 'No se pudo asignar el tecnico.');
    } finally {
      setSending(false);
    }
  };

  return (
    <section className="card">
      <h3 className="card__title">Asignacion</h3>
      <DataState
        loading={assignment.loading || (isAdministrator && technician.loading)}
        error={assignment.error || (isAdministrator ? technician.error : '')}
        empty={false}
      >
        <div style={{ marginBottom: '1rem', padding: '0.75rem', background: '#f8fafc', border: '1px solid #e2e8f0', borderRadius: '4px' }}>
          <strong>Técnico asignado actualmente:</strong>{' '}
          {assignment.data?.isActive ? (assignedTech ? `${assignedTech.fullName} (${assignedTech.specialty})` : 'Técnico asignado a la orden') : 'Sin técnico asignado'}
        </div>

        {isDelivered ? (
          <p className="state-message">La orden ya fue entregada. La asignación está cerrada.</p>
        ) : isAdministrator ? (
          <>
            <form onSubmit={submit} noValidate>
              <ErrorBanner message={error} />
              <SuccessBanner message={confirmation} />
              <div className="field">
                <label className="field__label" htmlFor="technicianId">
                  Tecnico
                </label>
                <select
                  className="field__input"
                  id="technicianId"
                  value={technicianId}
                  onChange={(event) => setTechnicianId(event.target.value)}
                  required
                >
                  <option value="">Seleccione un tecnico</option>
                  {(technician.data ?? []).map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.fullName + (item.id === assignment.data?.technicianId ? ' (actual)' : item.busy ? ' (ocupado)' : '')}
                    </option>
                  ))}
                </select>
              </div>
              <button type="submit" className="button button--primary" disabled={sending}>
                {sending ? 'Asignando...' : 'Asignar tecnico'}
              </button>
            </form>
            <div className="table-scroll">
              <table className="data-table">
                <thead>
                  <tr>
                    <th scope="col">Tecnico</th>
                    <th scope="col">Especialidad</th>
                    <th scope="col">Estado</th>
                    <th scope="col">Orden activa</th>
                  </tr>
                </thead>
                <tbody>
                  {(technician.data ?? []).map((item) => (
                    <tr key={item.id}>
                      <td>{item.fullName}</td>
                      <td>{item.specialty}</td>
                      <td>
                        <AvailabilityBadge busy={item.busy} />
                      </td>
                      <td>
                        {item.id === assignment.data?.technicianId
                          ? 'Asignado a esta orden'
                          : item.activeOrderNumber || 'Sin orden'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </>
        ) : (
          <p className="state-message">Solo el jefe de taller asigna tecnicos.</p>
        )}
      </DataState>
    </section>
  );
}
