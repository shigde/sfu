import {Component, OnInit} from '@angular/core';
import {Stream, StreamService, SpaceService} from '@shigde/core';
import {map, Observable, of} from 'rxjs';
import {v4 as uuidv4} from 'uuid';


@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent implements OnInit {
  streams$: Observable<Stream[]> = of([]);
  streamMap = new Map<string, Stream[]>();

  constructor(
    private spaceService: SpaceService,
    private streamService: StreamService
  ) {
  }

  ngOnInit(): void {
    this.streams$ = of([
      {uuid: uuidv4(), title: 'Test1', user: 'user@test.de'},
      {uuid: uuidv4(), title: 'Test2', user: 'user@test.de'},
      {uuid: uuidv4(), title: 'Test3', user: 'user@test.de'},
      {uuid: uuidv4(), title: 'Test4', user: 'user@test.de'},
      {uuid: uuidv4(), title: 'Test5', user: 'user@test.de'},
      {uuid: uuidv4(), title: 'Test6', user: 'user@test.de'},
      {uuid: uuidv4(), title: 'Test7', user: 'user@test.de'},
      {uuid: uuidv4(), title: 'Test8', user: 'user@test.de'},
      {uuid: uuidv4(), title: 'Test9', user: 'user@test.de'},
      {uuid: uuidv4(), title: 'Test10', user: 'user@test.de'},
      {uuid: uuidv4(), title: 'Test11', user: 'user@test.de'},
    ]);
  }

  getStreams(): void {
    // this.streamService.getStreams('live_stream_channel@localhost:9000')
    //   .subscribe(streams => this.streams = streams);
  }

  getSpaces(): void {
    this.spaceService.getSpaces()
      .pipe(map((spaces) => spaces.slice(1, 5)))
      .subscribe(spaces => {
        spaces.forEach(space => {
          this.streamService.getStreams(space.id).subscribe((streams) => {
            this.streamMap.set(space.id, streams);
          });
        });
      });
  }
}
