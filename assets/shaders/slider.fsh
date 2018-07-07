#version 330

uniform sampler2D tex;
uniform vec4 col_border;
uniform vec4 col_border1;

in vec2 tex_coord;
out vec4 color;
void main()
{
    vec4 in_color = texture2D(tex, tex_coord);

	color = in_color*mix(col_border1, col_border, smoothstep(45.0/512, 60.0/512, tex_coord.x));
}