import { getAppEvents } from '@grafana/runtime';
import { AppEvents, EventBus } from '@grafana/data';

type NotificatorType = {
  error: (title: string, message: string) => void;
  success: (title: string, message: string) => void;
  warning: (title: string, message: string) => void;
};

export class Notificator implements NotificatorType {
  private static instance: Notificator;
  private readonly appEvents: EventBus;
  private readonly ERROR_TYPE = AppEvents.alertError.name;
  private readonly SUCCESS_TYPE = AppEvents.alertSuccess.name;
  private readonly WARNING_TYPE = AppEvents.alertWarning.name;

  constructor() {
    this.appEvents = getAppEvents();
  }

  public static getInstance(): Notificator {
    Notificator.instance ||= new Notificator();

    return Notificator.instance;
  }

  public error(title: string, message?: string, ...rest: any[]) {
    this.publish(this.ERROR_TYPE, title, message, ...rest);
  }

  public success(title: string, message?: string, ...rest: any[]) {
    this.publish(this.SUCCESS_TYPE, title, message, ...rest);
  }

  public warning(title: string, message?: string, ...rest: any[]) {
    this.publish(this.WARNING_TYPE, title, message, ...rest);
  }

  private publish(type: string, title: string, message?: string, ...rest: any[]) {
    this.appEvents.publish({
      type,
      payload: [title, message, ...rest],
    });
  }
}
