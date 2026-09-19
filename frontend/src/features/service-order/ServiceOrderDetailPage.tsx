/* Order detail: the header with the status and the advance buttons, plus the
   four panels that hang from the order. Each panel is its own component. */
import { useState } from 'react';
import { useParams } from 'react-router-dom';

import { ApiError } from '../../services/api_client';
import {
  advanceServiceOrder,
  findAssignment,
  findServiceOrder,
} from '../../services/service_order_service';
import { listTechnician } from '../../services/technician_service';
import { AssignmentPanel } from './AssignmentPanel';
import { DiagnosticPanel } from './DiagnosticPanel';
import { InterventionPanel } from './InterventionPanel';
import { StatusHistoryPanel } from './StatusHistoryPanel';
import { DataState, ErrorBanner, SuccessBanner } from '../../shared/DataState';
import { StatusBadge } from '../../shared/StatusBadge';
import { useAsyncData } from '../../shared/useAsyncData';
import { useSession, useToken } from '../../shared/SessionContext';
import { formatDateTime, nextStatus, statusLabel } from '../../shared/format';

export function ServiceOrderDetailPage() {
  const { serviceOrderId = '' } = useParams();
  const token = useToken();
  const { session, isAdministrator } = useSession();
  const order = useAsyncData(() => findServiceOrder(token, serviceOrderId), [token, serviceOrderId]);
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
  const technicians = useAsyncData(() => listTechnician(token), [token]);

  const [actionError, setActionError] = useState('');
  const [confirmation, setConfirmation] = useState('');
  const [sending, setSending] = useState(false);
  const [historyToken, setHistoryToken] = useState(0);

  const current = order.data;
  const following = current ? nextStatus(current.status) : null;
  const isDelivered = current?.status === 'DELIVERED';
  const isTechnician = session?.role === 'TECHNICIAN';

  const myTechnician = (technicians.data ?? []).find((t) => t.userId === session?.userId);
  const isAssignedToMe = Boolean(
    assignment.data && myTechnician && assignment.data.technicianId === myTechnician.id,
  );
  const canEdit = !isDelivered && (!isTechnician || isAssignedToMe);
  const canAdvance = Boolean(following && !isDelivered && (isAdministrator || (isTechnician && isAssignedToMe)));

  const advance = async () => {
    if (!following) {
      return;
    }
    setActionError('');
    setConfirmation('');
    setSending(true);
    try {
      await advanceServiceOrder(token, serviceOrderId, following);
      setConfirmation('Estado actualizado a ' + statusLabel(following) + '.');
      order.reload();
      assignment.reload();
      technicians.reload();
      setHistoryToken((previous) => previous + 1);
    } catch (failure) {
      setActionError(
        failure instanceof ApiError ? failure.message : 'No se pudo cambiar el estado.',
      );
    } finally {
      setSending(false);
    }
  };

  return (
    <section>
      <h2 className="screen-title">Detalle de la orden</h2>
      <DataState loading={order.loading} error={order.error} empty={!current}>
        <section className="card">
          <h3 className="card__title">
            {current?.orderNumber} {current ? <StatusBadge status={current.status} /> : null}
          </h3>
          <p>
            <strong>Falla reportada:</strong> {current?.reportedFailure}
          </p>
          <p className="timeline__date">Ingreso: {formatDateTime(current?.receivedAt ?? '')}</p>
          <ErrorBanner message={actionError} />
          <SuccessBanner message={confirmation} />
          {canAdvance ? (
            <button
              type="button"
              className="button button--primary"
              onClick={advance}
              disabled={sending}
            >
              {sending ? 'Actualizando...' : 'Marcar ' + statusLabel(following!).toLowerCase()}
            </button>
          ) : null}
          {isDelivered ? <p className="state-message">La orden ya fue entregada.</p> : null}
          {isTechnician && !isAssignedToMe && !isDelivered ? (
            <p className="state-message">Esta orden no está asignada a su usuario. Solo tiene permisos de lectura.</p>
          ) : null}
        </section>

        <div className="panel-stack">
          <AssignmentPanel
            serviceOrderId={serviceOrderId}
            isDelivered={isDelivered}
            onChange={() => {
              order.reload();
              assignment.reload();
              technicians.reload();
            }}
          />
          <DiagnosticPanel
            serviceOrderId={serviceOrderId}
            canEdit={canEdit}
            isDelivered={isDelivered}
            isAssignedToMe={isAssignedToMe}
            onChange={() => order.reload()}
          />
          <InterventionPanel
            serviceOrderId={serviceOrderId}
            canEdit={canEdit}
            isDelivered={isDelivered}
            isAssignedToMe={isAssignedToMe}
            onChange={() => order.reload()}
          />
          <StatusHistoryPanel serviceOrderId={serviceOrderId} refreshToken={historyToken} />
        </div>
      </DataState>
    </section>
  );
}
