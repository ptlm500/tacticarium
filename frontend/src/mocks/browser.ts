import { setupWorker } from 'msw/browser';
import { restHandlers } from './handlers/rest';
import { wsHandlers } from './handlers/ws';

export const worker = setupWorker(...restHandlers, ...wsHandlers);
