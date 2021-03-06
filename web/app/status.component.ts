import { Component } from '@angular/core';

import { StatusService } from './status.service'

@Component({
  moduleId: module.id,
  selector: 'dmx-status',
  host: { class: 'status' },
  templateUrl: 'status.component.html',
})
export class StatusComponent  {
  constructor(private status: StatusService) { }
}
