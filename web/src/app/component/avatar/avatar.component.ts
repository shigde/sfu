import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-avatar',
  standalone: true,
  imports: [],
  templateUrl: './avatar.component.html',
  styleUrl: './avatar.component.scss'
})
export class AvatarComponent implements OnInit{
  @Input() name = 'dddd';
  @Input() size = '60';
  imageSrc = '';
  private colors = [
    '#81b997',
    '#59a5da',
    '#ffb3ba',
    '#f3bd8f',
    '#fd71c3',
    '#836d55',
    '#fdfd66',
    '#68ff8b',
    '#bae1ff', '#1abc9c', '#2ecc71', '#3498db', '#9b59b6', '#34495e', '#16a085', '#27ae60', '#2980b9', '#8e44ad', '#2c3e50',
    '#f1c40f', '#e67e22', '#e74c3c', '#ecf0f1', '#95a5a6', '#f39c12', '#d35400', '#c0392b', '#bdc3c7', '#7f8c8d'
  ];

  constructor() {
  }

  ngOnInit(): void {
    this.imageSrc = this.letterAvatar(this.name, Number(this.size));
  }

  private letterAvatar(name: string, size: number): string {
    const nameSplit = String(name).toUpperCase().split(' ');
    let initials;

    if (nameSplit.length === 1) {
      initials = nameSplit[0] ? nameSplit[0].charAt(0) : '?';
    } else {
      initials = nameSplit[0].charAt(0) + nameSplit[1].charAt(0);
    }

    if (window.devicePixelRatio) {
      size = (size * window.devicePixelRatio);
    }

    const charIndex = (initials === '?' ? 72 : initials.charCodeAt(0)) - 64;
    const colourIndex = charIndex % 29;
    const canvas = document.createElement('canvas');
    canvas.width = size;
    canvas.height = size;

    const context = canvas.getContext('2d');

    if (context) {
      context.fillStyle = this.colors[colourIndex - 1];
      context.fillRect(0, 0, canvas.width, canvas.height);
      context.font = Math.round(canvas.width / 2) + 'px Arial';
      context.textAlign = 'center';
      context.fillStyle = '#FFF';
      context.fillText(initials, size / 2, size / 1.5);

      return canvas.toDataURL();
    }

    return '';
  }

  transform(): void {
    Array.prototype.forEach.call(document.querySelectorAll('img[avatar]'), (img) => {
      const name = img.getAttribute('avatar');
      img.src = this.letterAvatar(name, img.getAttribute('width'));
      img.removeAttribute('avatar');
      img.setAttribute('alt', name);
    });
  }
}
