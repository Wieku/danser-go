#version 330

uniform sampler2DArray tex;
uniform sampler2DArray tex2;
uniform float power;

in vec2 tex_coord;
out vec4 color;

void main()
{
    vec4 in_color = texture(tex, vec3(tex_coord, 0));
    vec4 in_color2 = texture(tex2, vec3(tex_coord, 0));
	color = in_color + in_color2 * power;
}